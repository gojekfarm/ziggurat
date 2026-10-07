package otelemetry

import (
	"context"
	"testing"

	"github.com/gojekfarm/ziggurat/v2"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// withRecordingTracerProvider registers a TracerProvider backed by an in-memory span
// recorder for the duration of the test, and returns it so the test can inspect ended
// spans. Restores whatever TracerProvider was previously registered afterward, since
// otel's TracerProvider is process-global state.
func withRecordingTracerProvider(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })
	return sr
}

func TestTraceHandler_StartsSpanWithExpectedAttributes(t *testing.T) {
	sr := withRecordingTracerProvider(t)

	handler := ziggurat.HandlerFunc(func(ctx context.Context, e *ziggurat.Event) {})
	wrapped := TraceHandler("ziggurat-test")(handler)

	event := &ziggurat.Event{RoutingPath: "group/topic/0", EventType: "kafka"}
	wrapped.Handle(context.Background(), event)

	spans := sr.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 ended span, got %d", len(spans))
	}

	span := spans[0]
	if span.Name() != event.RoutingPath {
		t.Errorf("expected span name %q, got %q", event.RoutingPath, span.Name())
	}
	if span.Status().Code == codes.Error {
		t.Errorf("expected span status to not be Error for a successful handle, got: %v", span.Status())
	}

	attrs := map[string]string{}
	for _, a := range span.Attributes() {
		attrs[string(a.Key)] = a.Value.AsString()
	}
	if attrs["messaging.system"] != "kafka" {
		t.Errorf("expected messaging.system=kafka attribute, got %v", attrs)
	}
	if attrs["messaging.destination"] != event.RoutingPath {
		t.Errorf("expected messaging.destination=%q attribute, got %v", event.RoutingPath, attrs)
	}
}

// TestTraceHandler_PanicIsRecordedNotSwallowed mirrors the equivalent test in
// mw/prometheus: tracing must observe a panic (record it as a span error) but never
// swallow it.
func TestTraceHandler_PanicIsRecordedNotSwallowed(t *testing.T) {
	sr := withRecordingTracerProvider(t)

	handler := ziggurat.HandlerFunc(func(ctx context.Context, e *ziggurat.Event) {
		panic("deliberate test panic")
	})
	wrapped := TraceHandler("ziggurat-test")(handler)

	event := &ziggurat.Event{RoutingPath: "group/topic/0", EventType: "kafka"}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected the panic to propagate out of TraceHandler, but it was swallowed")
		}
		if r != "deliberate test panic" {
			t.Errorf("expected the original panic value to be preserved through re-panic, got: %v", r)
		}

		spans := sr.Ended()
		if len(spans) != 1 {
			t.Fatalf("expected 1 ended span, got %d", len(spans))
		}
		if spans[0].Status().Code != codes.Error {
			t.Errorf("expected span status Error after a panic, got: %v", spans[0].Status())
		}
		events := spans[0].Events()
		if len(events) == 0 {
			t.Fatal("expected the panic to be recorded as a span event (via RecordError), got none")
		}
	}()

	wrapped.Handle(context.Background(), event)
}

func TestHeadersFrom(t *testing.T) {
	t.Run("nil metadata yields empty carrier", func(t *testing.T) {
		event := &ziggurat.Event{}
		got := headersFrom(event)
		if len(got) != 0 {
			t.Errorf("expected empty carrier, got %v", got)
		}
	})

	t.Run("extracts existing headers map", func(t *testing.T) {
		event := &ziggurat.Event{
			Metadata: map[string]interface{}{
				ziggurat.MetadataKeyHeaders: map[string]string{"traceparent": "00-abc-def-01"},
			},
		}
		got := headersFrom(event)
		if got.Get("traceparent") != "00-abc-def-01" {
			t.Errorf("expected traceparent to be extracted, got %v", got)
		}
	})
}
