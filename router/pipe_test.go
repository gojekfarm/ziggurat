package router

import (
	"context"
	"testing"

	"github.com/gojekfarm/ziggurat"
)

func TestPipeHandlers(t *testing.T) {
	mw1 := func(next ziggurat.Handler) ziggurat.Handler {
		return ziggurat.HandlerFunc(func(ctx context.Context, messageEvent ziggurat.Event) error {
			me := ziggurat.CreateMockEvent()
			me.ValueFunc = func() []byte {
				return []byte("foo")
			}
			return next.HandleEvent(ctx, me)
		})
	}
	mw2 := func(next ziggurat.Handler) ziggurat.Handler {
		return ziggurat.HandlerFunc(func(ctx context.Context, messageEvent ziggurat.Event) error {
			byteValue := append(messageEvent.Value(), []byte("-bar")...)
			me := ziggurat.CreateMockEvent()
			me.ValueFunc = func() []byte {
				return byteValue
			}
			return next.HandleEvent(ctx, me)
		})
	}
	actualHandler := ziggurat.HandlerFunc(func(ctx context.Context, event ziggurat.Event) error {
		if string(event.Value()) != "foo-bar" {
			t.Errorf("expected message to be %s,but got %s", "foo-bar", string(event.Value()))
		}
		return nil
	})
	finalHandler := PipeHandlers(mw1, mw2)(actualHandler)
	finalHandler.HandleEvent(context.Background(), ziggurat.MockEvent{})
}
