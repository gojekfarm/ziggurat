package event

import (
	"context"

	"github.com/gojekfarm/ziggurat/v2"
	"go.opentelemetry.io/otel/trace"
)

// Logger logs every event it sees. Put otelemetry.TraceHandler before it in the
// ziggurat.Use chain to get trace_id/span_id on each line.
func Logger(l ziggurat.StructuredLogger) func(handler ziggurat.Handler) ziggurat.Handler {
	return func(handler ziggurat.Handler) ziggurat.Handler {
		f := func(ctx context.Context, event *ziggurat.Event) {
			kvs := map[string]interface{}{
				"path":               event.RoutingPath,
				"producer-timestamp": event.ProducerTimestamp,
				"received-timestamp": event.ReceivedTimestamp,
				"value-length":       len(event.Value),
				"event-type":         event.EventType,
				"logger-type":        "event.logger.middleware",
			}

			for k, v := range event.Metadata {
				kvs[k] = v
			}

			// Correlate this log line with its trace. Only present when a valid span is
			// active (i.e. otelemetry.TraceHandler is earlier in the middleware chain and
			// tracing is configured), so it costs nothing otherwise.
			if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
				kvs["trace_id"] = sc.TraceID().String()
				kvs["span_id"] = sc.SpanID().String()
			}

			l.Info("event received", kvs)
			handler.Handle(ctx, event)
		}
		return ziggurat.HandlerFunc(f)
	}
}
