package router

import (
	"context"
	"github.com/gojekfarm/ziggurat/mock"
	"testing"

	"github.com/gojekfarm/ziggurat"
)

func TestPipeHandlers(t *testing.T) {
	mw1 := func(next ziggurat.Handler) ziggurat.Handler {
		return ziggurat.HandlerFunc(func(ctx context.Context, messageEvent ziggurat.Event) error {
			me := mock.CreateMockEvent()
			me.ValueFunc = func() []byte {
				return []byte("foo")
			}
			return next.Handle(ctx, me)
		})
	}
	mw2 := func(next ziggurat.Handler) ziggurat.Handler {
		return ziggurat.HandlerFunc(func(ctx context.Context, messageEvent ziggurat.Event) error {
			byteValue := append(messageEvent.Value(), []byte("-bar")...)
			me := mock.CreateMockEvent()
			me.ValueFunc = func() []byte {
				return byteValue
			}
			return next.Handle(ctx, me)
		})
	}
	actualHandler := ziggurat.HandlerFunc(func(ctx context.Context, event ziggurat.Event) error {
		if string(event.Value()) != "foo-bar" {
			t.Errorf("expected message to be %s,but got %s", "foo-bar", string(event.Value()))
		}
		return nil
	})
	finalHandler := PipeHandlers(mw1, mw2)(actualHandler)
	finalHandler.Handle(context.Background(), mock.Event{})
}
