package rabbitmq

import (
	"context"
	"strings"
	"testing"

	"github.com/gojekfarm/ziggurat/v2"

	"github.com/makasim/amqpextra/publisher"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// TestPublishInternal_InjectsTraceContextWhenSpanActive proves the Phase D requirement
// directly: when ctx carries an active, sampled span, publishInternal's outgoing message
// headers carry a traceparent - so a retry/DLQ republish stays part of the same trace as
// the consume that triggered it, rather than starting a disconnected new one.
//
// publish_test.go's existing assertions (exact Headers: {"retry-origin": "ziggurat-go"})
// still pass unmodified because those tests use context.Background() with no
// TracerProvider/propagator configured in this test binary, where trace injection is a
// documented no-op by the OpenTelemetry spec - this test is what actually exercises the
// non-no-op path.
func TestPublishInternal_InjectsTraceContextWhenSpanActive(t *testing.T) {
	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
	})

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}))

	ctx, span := otel.Tracer("test").Start(context.Background(), "test-span")
	defer span.End()

	p := mockAMQPPublisher{}
	p.On("Publish", mock.Anything).Return(nil)

	event := &ziggurat.Event{}
	if err := publishInternal(ctx, &p, "foo", 1, "100", event); err != nil {
		t.Fatalf("publishInternal returned error: %v", err)
	}

	if len(p.Calls) != 1 {
		t.Fatalf("expected Publish to have been called exactly once, got %d calls", len(p.Calls))
	}
	msg, ok := p.Calls[0].Arguments[0].(publisher.Message)
	if !ok {
		t.Fatalf("expected Publish's argument to be a publisher.Message, got %T", p.Calls[0].Arguments[0])
	}

	tp_, ok := msg.Publishing.Headers["traceparent"]
	if !ok {
		t.Fatalf("expected a traceparent header to be injected, got headers: %v", msg.Publishing.Headers)
	}
	traceparent, ok := tp_.(string)
	if !ok || !strings.HasPrefix(traceparent, "00-") {
		t.Errorf("expected a W3C-format traceparent string (00-...), got: %v", tp_)
	}
	if msg.Publishing.Headers["retry-origin"] != "ziggurat-go" {
		t.Errorf("expected retry-origin header to still be present alongside traceparent, got: %v", msg.Publishing.Headers)
	}
}
