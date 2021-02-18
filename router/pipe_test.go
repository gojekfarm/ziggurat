package router

import (
	"context"
	"github.com/gojekfarm/ziggurat"
	"testing"
)

func TestPipeHandlers(t *testing.T) {
	mw1 := func(next ziggurat.Handler) ziggurat.Handler {
		return ziggurat.HandlerFunc(func(messageEvent ziggurat.Event, ctx context.Context) error {
			me := ziggurat.CreateMockEvent()
			me.ValueFunc = func() []byte {
				return []byte("foo")
			}
			return next.HandleEvent(me, ctx)
		})
	}
	mw2 := func(next ziggurat.Handler) ziggurat.Handler {
		return ziggurat.HandlerFunc(func(messageEvent ziggurat.Event, ctx context.Context) error {
			byteValue := append(messageEvent.Value(), []byte("-bar")...)
			me := ziggurat.CreateMockEvent()
			me.ValueFunc = func() []byte {
				return byteValue
			}
			return next.HandleEvent(me, ctx)
		})
	}
	actualHandler := ziggurat.HandlerFunc(func(event ziggurat.Event, ctx context.Context) error {
		if string(event.Value()) != "foo-bar" {
			t.Errorf("expected message to be %s,but got %s", "foo-bar", string(event.Value()))
		}
		return nil
	})
	finalHandler := PipeHandlers(mw1, mw2)(actualHandler)
	finalHandler.HandleEvent(ziggurat.MockEvent{}, context.Background())
}
