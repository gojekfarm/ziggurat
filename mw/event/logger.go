package event

import (
	"context"

	"github.com/gojekfarm/ziggurat"
)

func Logger(l ziggurat.StructuredLogger) func(handler ziggurat.Handler) ziggurat.Handler {
	return func(handler ziggurat.Handler) ziggurat.Handler {
		f := func(ctx context.Context, event *ziggurat.Event) error {
			l.Info("received message", map[string]interface{}{
				"path":               event.Path,
				"producer-timestamp": event.ProducerTimestamp,
				"received-timestamp": event.ReceivedTimestamp,
				"value-length":       len(event.Value),
				"event-type":         event.EventType,
			})
			return handler.Handle(ctx, event)
		}
		return ziggurat.HandlerFunc(f)
	}
}
