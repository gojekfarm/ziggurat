package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/gojekfarm/ziggurat/v2"
	"github.com/gojekfarm/ziggurat/v2/mw/otelemetry"
	"github.com/gojekfarm/ziggurat/v2/mw/prometheus"
	"github.com/makasim/amqpextra/publisher"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

// resetOTel puts no-op globals back. Restoring the "previous" globals does not work:
// otel's default delegating provider/propagator permanently delegate to the first real
// one ever set, so a restored default would keep injecting traceparent headers into later
// tests that assert exact headers.
func resetOTel() {
	otel.SetTracerProvider(tracenoop.NewTracerProvider())
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())
}

func installRecorder(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	t.Cleanup(resetOTel)
	rec := tracetest.NewSpanRecorder()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()), sdktrace.WithSpanProcessor(rec)))
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return rec
}

func TestPublishObserved_SpanIsProducerAndItsContextIsInjected(t *testing.T) {
	rec := installRecorder(t)
	ctx, parent := otel.Tracer("t").Start(context.Background(), "consume")

	p := mockAMQPPublisher{}
	p.On("Publish", mock.Anything).Return(nil)
	if err := publishObserved(ctx, &p, "obs-q-ok", outcomeRetry, 2, publisher.Message{}); err != nil {
		t.Fatal(err)
	}
	parent.End()

	var pub sdktrace.ReadOnlySpan
	for _, s := range rec.Ended() {
		if s.Name() == "obs-q-ok publish" {
			pub = s
		}
	}
	if pub == nil {
		t.Fatal("no publish span recorded")
	}
	if pub.SpanKind() != trace.SpanKindProducer || pub.Parent().SpanID() != parent.SpanContext().SpanID() {
		t.Errorf("publish span must be a producer child of the consume span")
	}
	attrs := map[string]string{}
	for _, kv := range pub.Attributes() {
		attrs[string(kv.Key)] = kv.Value.Emit()
	}
	if attrs["ziggurat.retry.outcome"] != "retry" || attrs["ziggurat.retry.count"] != "2" {
		t.Errorf("unexpected attrs %v", attrs)
	}

	// The header must carry the PUBLISH span's id, so the next consume is its child.
	msg := p.Calls[0].Arguments[0].(publisher.Message)
	tp, _ := msg.Publishing.Headers["traceparent"].(string)
	if !strings.Contains(tp, pub.SpanContext().SpanID().String()) {
		t.Errorf("traceparent %q must reference the publish span %s", tp, pub.SpanContext().SpanID())
	}
	if msg.Publishing.Headers["retry-origin"] != "ziggurat-go" {
		t.Error("retry-origin header missing")
	}
}

func TestPublishObserved_ErrorClassification(t *testing.T) {
	cases := []struct {
		queue  string
		err    error
		reason string
	}{
		{"obs-q-nack", errors.New("confirmation: nack"), prometheus.PublishErrNack},
		{"obs-q-other", errors.New("connection closed"), prometheus.PublishErrPublish},
	}
	for _, c := range cases {
		p := mockAMQPPublisher{}
		p.On("Publish", mock.Anything).Return(c.err)
		if err := publishObserved(context.Background(), &p, c.queue, outcomeDLQ, 0, publisher.Message{}); err == nil {
			t.Fatalf("%s: expected the error to be returned", c.queue)
		}
		if got := testutil.ToFloat64(prometheus.RetryPublishErrorsCounter.WithLabelValues(c.queue, c.reason)); got != 1 {
			t.Errorf("%s: errors{reason=%s} = %v, want 1", c.queue, c.reason, got)
		}
		if n := testutil.CollectAndCount(prometheus.RetryPublishDurationHistogram); n == 0 {
			t.Errorf("%s: expected a publish duration observation", c.queue)
		}
	}
}

func TestPublishInternal_CountsPublishedOnlyOnSuccess(t *testing.T) {
	const queue = "obs-q-count"
	p := mockAMQPPublisher{}
	p.On("Publish", mock.Anything).Return(errors.New("confirmation: nack")).Once()
	p.On("Publish", mock.Anything).Return(nil).Once()

	if err := publishInternal(context.Background(), &p, queue, 5, "100", &ziggurat.Event{}); err == nil {
		t.Fatal("expected nack error")
	}
	if got := testutil.ToFloat64(prometheus.RetryPublishedCounter.WithLabelValues(queue)); got != 0 {
		t.Errorf("a failed publish must not count as published, got %v", got)
	}
	if err := publishInternal(context.Background(), &p, queue, 5, "100", &ziggurat.Event{}); err != nil {
		t.Fatal(err)
	}
	if got := testutil.ToFloat64(prometheus.RetryPublishedCounter.WithLabelValues(queue)); got != 1 {
		t.Errorf("published = %v, want 1", got)
	}
}

func TestPublishInternal_DLQCountedWhenRetriesExhausted(t *testing.T) {
	const queue = "obs-q-dlq"
	p := mockAMQPPublisher{}
	p.On("Publish", mock.Anything).Return(nil)
	if err := publishInternal(context.Background(), &p, queue, 0, "100", &ziggurat.Event{}); err != nil {
		t.Fatal(err)
	}
	if got := testutil.ToFloat64(prometheus.DLQPublishedCounter.WithLabelValues(queue)); got != 1 {
		t.Errorf("dlq published = %v, want 1", got)
	}
	if got := testutil.ToFloat64(prometheus.RetryPublishedCounter.WithLabelValues(queue)); got != 0 {
		t.Errorf("retry published = %v, want 0", got)
	}
}

func TestReplayHeaders_AddsMarkerWithoutMutatingOriginal(t *testing.T) {
	orig := amqp.Table{"traceparent": "00-abc-def-01", "retry-origin": "ziggurat-go"}
	got := replayHeaders(orig)
	if got[otelemetry.HeaderReplay] != "true" || got["traceparent"] != "00-abc-def-01" {
		t.Errorf("unexpected headers %v", got)
	}
	if _, polluted := orig[otelemetry.HeaderReplay]; polluted {
		t.Error("the delivery's own header table must not be modified")
	}
	if len(replayHeaders(nil)) != 1 {
		t.Error("nil headers must still produce the marker")
	}
}

func TestDecodeDelivery_StampsQueueAndSurvivesRoundTrip(t *testing.T) {
	orig := ziggurat.Event{
		RoutingPath: "grp/orders/3",
		EventType:   "kafka",
		Metadata:    map[string]any{ziggurat.MetadataKeyKafkaTopic: "orders", ziggurat.MetadataKeyKafkaPartition: 3},
	}
	body, _ := json.Marshal(orig)

	e, err := decodeDelivery(body, amqp.Table{"traceparent": "x"}, "orders_instant_queue")
	if err != nil {
		t.Fatal(err)
	}
	if e.Metadata[ziggurat.MetadataKeyRabbitMQQueue] != "orders_instant_queue" {
		t.Error("queue key must be stamped on every consume")
	}
	if e.EventType != "kafka" {
		t.Error("the event's origin type must be preserved")
	}
	if h, _ := e.Metadata[ziggurat.MetadataKeyHeaders].(map[string]string); h["traceparent"] != "x" {
		t.Errorf("headers not captured: %v", e.Metadata)
	}

	// a stale queue key from a previous hop must be overwritten
	orig.Metadata[ziggurat.MetadataKeyRabbitMQQueue] = "old"
	body, _ = json.Marshal(orig)
	e, _ = decodeDelivery(body, nil, "new")
	if e.Metadata[ziggurat.MetadataKeyRabbitMQQueue] != "new" {
		t.Error("stale queue key was not overwritten")
	}

	if _, err := decodeDelivery([]byte("{bad"), nil, "q"); err == nil {
		t.Error("expected an error for an invalid body")
	}
}

func TestPublisherPoolInUseGauge(t *testing.T) {
	// only the accounting is exercised: get() on a non-empty pool and put() back
	pool := &publisherPool{pool: make(chan *publisher.Publisher, 1)}
	pool.pool <- nil
	before := testutil.ToFloat64(prometheus.RetryPubpoolInUseGauge)
	if _, err := pool.get(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := testutil.ToFloat64(prometheus.RetryPubpoolInUseGauge); got != before+1 {
		t.Errorf("in_use after get = %v, want %v", got, before+1)
	}
	// the pool has room again, so put() just returns the slot (no Close on the nil value)
	pool.put(nil)
	if got := testutil.ToFloat64(prometheus.RetryPubpoolInUseGauge); got != before {
		t.Errorf("in_use after put = %v, want %v", got, before)
	}
}
