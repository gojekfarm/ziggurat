package router

import (
	"context"
	"github.com/gojekfarm/ziggurat"
	"testing"
)

func TestPipeHandlers(t *testing.T) {
	mw1 := func(next ziggurat.Handler) ziggurat.Handler {
		return ziggurat.HandlerFunc(func(messageEvent ziggurat.Event) ziggurat.ProcessStatus {
			me := ziggurat.CreateMessageEvent([]byte("foo"), nil, context.Background())
			return next.HandleEvent(me)
		})
	}
	mw2 := func(next ziggurat.Handler) ziggurat.Handler {
		return ziggurat.HandlerFunc(func(messageEvent ziggurat.Event) ziggurat.ProcessStatus {
			byteValue := append(messageEvent.Value(), []byte("-bar")...)
			me := ziggurat.CreateMessageEvent(byteValue, nil, context.Background())
			return next.HandleEvent(me)
		})
	}
	actualHandler := ziggurat.HandlerFunc(func(event ziggurat.Event) ziggurat.ProcessStatus {
		if string(event.Value()) != "foo-bar" {
			t.Errorf("expected message to be %s,but got %s", "foo-bar", string(event.Value()))
		}
		return ziggurat.ProcessingSuccess
	})
	finalHandler := PipeHandlers(mw1, mw2)(actualHandler)
	finalHandler.HandleEvent(ziggurat.Message{})
}
