package event

import (
	"context"
	"github.com/gojekfarm/ziggurat/v2"
)

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

			l.Info("event received", kvs)
			handler.Handle(ctx, event)
		}
		return ziggurat.HandlerFunc(f)
	}
}
