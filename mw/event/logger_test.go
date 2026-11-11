package event

import (
	"context"
	"testing"

	"github.com/gojekfarm/ziggurat/v2"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type captureLogger struct {
	fields map[string]interface{}
}

func (c *captureLogger) Info(_ string, kvs ...map[string]interface{}) {
	if len(kvs) > 0 {
		c.fields = kvs[0]
	}
}
func (c *captureLogger) Fatal(string, error, ...map[string]interface{}) {}
func (c *captureLogger) Debug(string, ...map[string]interface{})        {}
func (c *captureLogger) Warn(string, ...map[string]interface{})         {}
func (c *captureLogger) Error(string, error, ...map[string]interface{}) {}

func TestLogger_AddsTraceIDsWhenSpanActive(t *testing.T) {
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	defer func() { _ = tp.Shutdown(context.Background()) }()
	ctx, span := tp.Tracer("t").Start(context.Background(), "s")
	defer span.End()

	cl := &captureLogger{}
	Logger(cl)(ziggurat.HandlerFunc(func(context.Context, *ziggurat.Event) {})).Handle(ctx, &ziggurat.Event{RoutingPath: "a/b/0"})

	if cl.fields["trace_id"] != span.SpanContext().TraceID().String() {
		t.Errorf("trace_id = %v, want %s", cl.fields["trace_id"], span.SpanContext().TraceID())
	}
	if cl.fields["span_id"] != span.SpanContext().SpanID().String() {
		t.Errorf("span_id = %v", cl.fields["span_id"])
	}
}

func TestLogger_OmitsTraceIDsWithoutSpan(t *testing.T) {
	cl := &captureLogger{}
	Logger(cl)(ziggurat.HandlerFunc(func(context.Context, *ziggurat.Event) {})).Handle(context.Background(), &ziggurat.Event{})
	if _, ok := cl.fields["trace_id"]; ok {
		t.Error("trace_id must not be logged when there is no active span")
	}
}
