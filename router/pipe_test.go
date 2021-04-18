package router

import (
	"context"
	"testing"
	"time"

	"github.com/gojekfarm/ziggurat"
)

func TestPipeHandlers(t *testing.T) {
	mw1 := func(next ziggurat.Handler) ziggurat.Handler {
		return ziggurat.HandlerFunc(func(ctx context.Context, messageEvent *ziggurat.Event) interface{} {
			me := ziggurat.Event{
				Headers:           nil,
				Value:             []byte("foo"),
				Key:               nil,
				Path:              "",
				ProducerTimestamp: time.Time{},
				ReceivedTimestamp: time.Time{},
				EventType:         "",
			}

			return next.Handle(ctx, &me)
		})
	}
	mw2 := func(next ziggurat.Handler) ziggurat.Handler {
		return ziggurat.HandlerFunc(func(ctx context.Context, messageEvent *ziggurat.Event) interface{} {
			byteValue := append(messageEvent.Value, []byte("-bar")...)
			me := ziggurat.Event{
				Headers:           nil,
				Value:             byteValue,
				Key:               nil,
				Path:              "",
				ProducerTimestamp: time.Time{},
				ReceivedTimestamp: time.Time{},
				EventType:         "",
			}

			return next.Handle(ctx, &me)
		})
	}
	actualHandler := ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) interface{} {
		if string(event.Value) != "foo-bar" {
			t.Errorf("expected message to be %s,but got %s", "foo-bar", string(event.Value))
		}
		return nil
	})
	finalHandler := PipeHandlers(mw1, mw2)(actualHandler)
	finalHandler.Handle(context.Background(), &ziggurat.Event{})
}
