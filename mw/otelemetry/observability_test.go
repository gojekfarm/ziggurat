package otelemetry

import (
	"context"
	"crypto/tls"
	"strings"
	"testing"

	"github.com/gojekfarm/ziggurat/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

// noopPropagator is a comparable propagator that does nothing (the composite
// propagator is a slice, which panics when compared, as setup_test.go does).
type noopPropagator struct{}

func (noopPropagator) Inject(context.Context, propagation.TextMapCarrier) {}
func (noopPropagator) Extract(ctx context.Context, _ propagation.TextMapCarrier) context.Context {
	return ctx
}
func (noopPropagator) Fields() []string { return nil }

// withRecorder installs a global TracerProvider that records every span, plus the W3C
// propagator, and restores the previous globals when the test ends.
func withRecorder(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	// no-op globals on cleanup: restoring the "previous" ones does not work because
	// otel's default delegating provider/propagator permanently delegate to the first
	// real one ever set.
	t.Cleanup(func() {
		otel.SetTracerProvider(tracenoop.NewTracerProvider())
		otel.SetTextMapPropagator(noopPropagator{})
	})
	rec := tracetest.NewSpanRecorder()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(rec),
	))
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return rec
}

func attrMap(s sdktrace.ReadOnlySpan) map[string]attribute.Value {
	m := map[string]attribute.Value{}
	for _, kv := range s.Attributes() {
		m[string(kv.Key)] = kv.Value
	}
	return m
}

func kafkaEvent() *ziggurat.Event {
	return &ziggurat.Event{
		RoutingPath: "grp/orders/3",
		EventType:   "kafka",
		Metadata: map[string]any{
			ziggurat.MetadataKeyKafkaTopic:     "orders",
			ziggurat.MetadataKeyKafkaPartition: 3,
		},
	}
}

func TestConsumeSpanInfo_Kafka(t *testing.T) {
	name, attrs := consumeSpanInfo(kafkaEvent())
	if name != "orders process" {
		t.Fatalf("span name must not contain the partition, got %q", name)
	}
	got := map[string]string{}
	for _, kv := range attrs {
		got[string(kv.Key)] = kv.Value.Emit()
	}
	want := map[string]string{
		"messaging.system":                      "kafka",
		"messaging.destination.name":            "orders",
		"messaging.kafka.consumer.group":        "grp",
		"messaging.kafka.destination.partition": "3",
		"messaging.operation":                   "process",
		"messaging.destination":                 "grp/orders/3",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("attr %s = %q, want %q", k, got[k], v)
		}
	}
}

func TestConsumeSpanInfo_RabbitMQRetryOfKafkaEvent(t *testing.T) {
	e := kafkaEvent()
	e.Metadata[ziggurat.MetadataKeyRabbitMQQueue] = "orders_instant_queue"
	e.Metadata[retryCountMetadataKey] = float64(2) // JSON round trip
	name, attrs := consumeSpanInfo(e)
	if name != "orders_instant_queue process" {
		t.Fatalf("got span name %q", name)
	}
	got := map[string]string{}
	for _, kv := range attrs {
		got[string(kv.Key)] = kv.Value.Emit()
	}
	if got["messaging.system"] != "rabbitmq" {
		t.Errorf("a RabbitMQ delivery must report messaging.system=rabbitmq, got %q", got["messaging.system"])
	}
	if got["ziggurat.event_origin"] != "kafka" || got["ziggurat.retry.count"] != "2" {
		t.Errorf("expected origin=kafka retry.count=2, got %v", got)
	}
}

func TestConsumeSpanInfo_FallsBackToRoutingPath(t *testing.T) {
	name, _ := consumeSpanInfo(&ziggurat.Event{RoutingPath: "custom/path", EventType: "x"})
	if name != "custom/path" {
		t.Fatalf("got %q", name)
	}
}

func TestTraceHandler_ConsumerKindAndPanicAttribute(t *testing.T) {
	rec := withRecorder(t)
	h := TraceHandler("t")(ziggurat.HandlerFunc(func(context.Context, *ziggurat.Event) { panic("boom") }))

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("panic must be re-raised")
			}
		}()
		h.Handle(context.Background(), kafkaEvent())
	}()

	spans := rec.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	s := spans[0]
	if s.SpanKind() != trace.SpanKindConsumer {
		t.Errorf("span kind = %v, want consumer", s.SpanKind())
	}
	if v := attrMap(s)["error.type"]; v.AsString() != "panic" {
		t.Errorf("expected error.type=panic, got %v", v)
	}
	if s.Status().Code.String() != "Error" {
		t.Errorf("status = %v", s.Status())
	}
}

func TestTraceHandler_ReplayStartsNewTraceLinkedToOriginal(t *testing.T) {
	rec := withRecorder(t)

	// an "original" trace whose context rides in the message headers
	origCtx, orig := otel.Tracer("t").Start(context.Background(), "original")
	carrier := HeaderCarrier{}
	InjectTraceContext(origCtx, carrier)
	orig.End()

	h := TraceHandler("t")(ziggurat.HandlerFunc(func(context.Context, *ziggurat.Event) {}))

	run := func(replay bool) sdktrace.ReadOnlySpan {
		hdr := HeaderCarrier{}
		for k, v := range carrier {
			hdr[k] = v
		}
		if replay {
			hdr[HeaderReplay] = "true"
		}
		e := kafkaEvent()
		e.Metadata[ziggurat.MetadataKeyHeaders] = map[string]string(hdr)
		before := len(rec.Ended())
		h.Handle(context.Background(), e)
		return rec.Ended()[before]
	}

	normal := run(false)
	if normal.Parent().TraceID() != orig.SpanContext().TraceID() || normal.SpanContext().TraceID() != orig.SpanContext().TraceID() {
		t.Error("a normal retry delivery must continue the original trace")
	}

	replayed := run(true)
	if replayed.SpanContext().TraceID() == orig.SpanContext().TraceID() {
		t.Error("a replay must start a NEW trace")
	}
	if replayed.Parent().IsValid() {
		t.Error("a replay span must be a root, not a child")
	}
	links := replayed.Links()
	if len(links) != 1 || links[0].SpanContext.TraceID() != orig.SpanContext().TraceID() {
		t.Errorf("expected one link to the original trace, got %v", links)
	}
}

func TestStartPublishSpan_IsProducerChildOfCaller(t *testing.T) {
	rec := withRecorder(t)
	ctx, parent := otel.Tracer("t").Start(context.Background(), "consume")
	_, span := StartPublishSpan(ctx, "orders", attribute.String("ziggurat.retry.outcome", "retry"))
	span.End()
	parent.End()

	var pub sdktrace.ReadOnlySpan
	for _, s := range rec.Ended() {
		if s.Name() == "orders publish" {
			pub = s
		}
	}
	if pub == nil {
		t.Fatal("publish span not recorded")
	}
	if pub.SpanKind() != trace.SpanKindProducer {
		t.Errorf("kind = %v", pub.SpanKind())
	}
	if pub.Parent().SpanID() != parent.SpanContext().SpanID() {
		t.Error("publish span must be a child of the consuming span")
	}
	if attrMap(pub)["ziggurat.retry.outcome"].AsString() != "retry" {
		t.Error("extra attributes must be applied")
	}
}

// ---------------- Setup ----------------

func TestSetupConfig_DefaultSamplerIsParentBasedRatio(t *testing.T) {
	var c setupConfig
	d := c.sdkSampler().Description()
	if !strings.Contains(d, "ParentBased") || !strings.Contains(d, "TraceIDRatioBased{0.1}") {
		t.Fatalf("unexpected default sampler: %s", d)
	}
}

func TestSetupConfig_SamplerOverrides(t *testing.T) {
	c := setupConfig{}
	WithSampleRatio(0.5)(&c)
	if d := c.sdkSampler().Description(); !strings.Contains(d, "TraceIDRatioBased{0.5}") {
		t.Errorf("WithSampleRatio not applied: %s", d)
	}
	WithSampler(sdktrace.AlwaysSample())(&c)
	if d := c.sdkSampler().Description(); d != "AlwaysOnSampler" {
		t.Errorf("WithSampler must win: %s", d)
	}
}

func TestSetupConfig_ResourceCarriesVersionEnvironmentAndExtras(t *testing.T) {
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "k8s.pod.name=pod-1")
	var c setupConfig
	WithServiceVersion("v9")(&c)
	WithEnvironment("staging")(&c)
	WithResourceAttributes(attribute.String("ziggurat.consumer_group", "grp"))(&c)

	res, err := c.buildResource(context.Background(), "svc")
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, kv := range res.Attributes() {
		got[string(kv.Key)] = kv.Value.Emit()
	}
	for k, v := range map[string]string{
		"service.name": "svc", "service.version": "v9", "deployment.environment": "staging",
		"ziggurat.consumer_group": "grp", "k8s.pod.name": "pod-1",
	} {
		if got[k] != v {
			t.Errorf("resource %s = %q, want %q", k, got[k], v)
		}
	}
	if got["host.name"] == "" {
		t.Error("expected host.name from the host detector")
	}
}

func TestSetup_EnabledWithOptions(t *testing.T) {
	prevTP, prevProp := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
	})
	shutdown, err := Setup(context.Background(), "svc", "localhost:4317", true,
		WithSampleRatio(1),
		WithServiceVersion("v1"),
		WithTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12}),
		WithBatchOptions(sdktrace.WithMaxQueueSize(128)),
	)
	if err != nil {
		t.Fatalf("Setup with options failed: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()
	if otel.GetTracerProvider() == prevTP {
		t.Error("expected a new global TracerProvider")
	}
}
