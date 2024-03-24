package ziggurat

import (
	"context"
	"testing"
	"time"
)

func TestPipeHandlers(t *testing.T) {
	mw1 := func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, messageEvent *Event) error {
			me := Event{
				Value:             []byte("foo"),
				Key:               nil,
				ProducerTimestamp: time.Time{},
				ReceivedTimestamp: time.Time{},
				EventType:         "",
			}

			return next.Handle(ctx, &me)
		})
	}
	mw2 := func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, messageEvent *Event) error {
			byteValue := append(messageEvent.Value, []byte("-bar")...)
			me := Event{
				RoutingPath:       "",
				Value:             byteValue,
				Key:               nil,
				ProducerTimestamp: time.Time{},
				ReceivedTimestamp: time.Time{},
				EventType:         "",
			}

			return next.Handle(ctx, &me)
		})
	}
	actualHandler := HandlerFunc(func(ctx context.Context, event *Event) error {
		if string(event.Value) != "foo-bar" {
			t.Errorf("expected message to be %s,but got %s", "foo-bar", string(event.Value))
		}
		return nil
	})
	finalHandler := pipe(actualHandler, mw1, mw2)
	_ = finalHandler.Handle(context.Background(), &Event{})
}
