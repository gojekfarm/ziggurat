package prometheus

import (
	"context"
	"testing"

	"github.com/gojekfarm/ziggurat/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func kafkaEvent(path, topic string, partition int) *ziggurat.Event {
	return &ziggurat.Event{
		RoutingPath: path,
		Metadata: map[string]interface{}{
			ziggurat.MetadataKeyKafkaTopic:     topic,
			ziggurat.MetadataKeyKafkaPartition: partition,
		},
	}
}

func TestRouteInfo(t *testing.T) {
	tests := []struct {
		name             string
		event            *ziggurat.Event
		withoutPartition bool
		route, topic, gr string
	}{
		{"kafka full route", kafkaEvent("grp/orders/3", "orders", 3), false, "grp/orders/3", "orders", "grp"},
		{"kafka partition dropped", kafkaEvent("grp/orders/3", "orders", 3), true, "grp/orders", "orders", "grp"},
		{"group containing slash", kafkaEvent("team/grp/orders/3", "orders", 3), true, "team/grp/orders", "orders", "team/grp"},
		{"no metadata keeps full path", &ziggurat.Event{RoutingPath: "a/b/c"}, true, "a/b/c", "", ""},
		{"metadata does not match path", kafkaEvent("grp/other/3", "orders", 3), true, "grp/other/3", "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route, topic, group := routeInfo(tc.event, tc.withoutPartition)
			if route != tc.route || topic != tc.topic || group != tc.gr {
				t.Fatalf("got (%q,%q,%q), want (%q,%q,%q)", route, topic, group, tc.route, tc.topic, tc.gr)
			}
		})
	}
}

func TestHandlerMetrics_WithoutPartitionInRoute_SharesSeriesAcrossPartitions(t *testing.T) {
	h := HandlerMetrics(WithoutPartitionInRoute())(ziggurat.HandlerFunc(func(context.Context, *ziggurat.Event) {}))
	h.Handle(context.Background(), kafkaEvent("lbl-grp/lbl-topic/0", "lbl-topic", 0))
	h.Handle(context.Background(), kafkaEvent("lbl-grp/lbl-topic/1", "lbl-topic", 1))

	got := testutil.ToFloat64(HandlerEventsCounter.With(prometheus.Labels{
		RouteLabel: "lbl-grp/lbl-topic", OutcomeLabel: OutcomeSuccess, TopicLabel: "lbl-topic", GroupLabel: "lbl-grp",
	}))
	if got != 2 {
		t.Fatalf("expected both partitions to share one series with count 2, got %v", got)
	}
}

func TestRegister_IsIdempotent(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("second Register call panicked: %v", r)
		}
	}()
	Register()
	Register()
}

func TestSetAppVersion_ReplacesPreviousVersion(t *testing.T) {
	SetAppVersion("v1")
	SetAppVersion("v2")
	if n := testutil.CollectAndCount(AppInfoGauge); n != 1 {
		t.Fatalf("expected exactly one app_info series, got %d", n)
	}
	if got := testutil.ToFloat64(AppInfoGauge.WithLabelValues("v2")); got != 1 {
		t.Fatalf("expected app_info{version=v2} = 1, got %v", got)
	}
}

func TestHandlerMetrics_AttachesTraceExemplarWhenSampled(t *testing.T) {
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	defer func() { _ = tp.Shutdown(context.Background()) }()
	ctx, span := tp.Tracer("test").Start(context.Background(), "x")
	defer span.End()

	h := HandlerMetrics()(ziggurat.HandlerFunc(func(context.Context, *ziggurat.Event) {}))
	h.Handle(ctx, &ziggurat.Event{RoutingPath: "exemplar/path"})

	reg := prometheus.NewPedanticRegistry()
	if err := reg.Register(HandlerDurationHistogram); err != nil {
		t.Fatal(err)
	}
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	want := span.SpanContext().TraceID().String()
	for _, mf := range mfs {
		for _, m := range mf.GetMetric() {
			for _, b := range m.GetHistogram().GetBucket() {
				if ex := b.GetExemplar(); ex != nil {
					for _, lp := range ex.GetLabel() {
						if lp.GetName() == "trace_id" && lp.GetValue() == want {
							return
						}
					}
				}
			}
		}
	}
	t.Fatalf("no exemplar with trace_id=%s found on the duration histogram", want)
}
