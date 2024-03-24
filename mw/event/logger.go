package event

import (
	"context"

	"github.com/gojekfarm/ziggurat"
)

func Logger(l ziggurat.StructuredLogger) func(handler ziggurat.Handler) ziggurat.Handler {
	return func(handler ziggurat.Handler) ziggurat.Handler {
		f := func(ctx context.Context, event *ziggurat.Event) error {
			kvs := map[string]interface{}{
				"path":               event.RoutingPath,
				"producer-timestamp": event.ProducerTimestamp,
				"received-timestamp": event.ReceivedTimestamp,
				"value-length":       len(event.Value),
				"event-type":         event.EventType,
			}

			for k, v := range event.Metadata {
				kvs[k] = v
			}

			l.Info("processing message", kvs)
			return handler.Handle(ctx, event)
		}
		return ziggurat.HandlerFunc(f)
	}
}
