package ziggurat

import (
	"context"
	"testing"
)

func TestPipeHandlers(t *testing.T) {
	mw1 := func(next Handler) Handler {
		return HandlerFunc(func(messageEvent Message, ctx context.Context) ProcessStatus {
			messageEvent.Value = []byte("foo")
			return next.HandleMessage(messageEvent, ctx)
		})
	}
	mw2 := func(next Handler) Handler {
		return HandlerFunc(func(messageEvent Message, ctx context.Context) ProcessStatus {
			messageEvent.Value = append(messageEvent.Value, []byte("-bar")...)
			return next.HandleMessage(messageEvent, ctx)
		})
	}
	actualHandler := HandlerFunc(func(event Message, ctx context.Context) ProcessStatus {
		if string(event.Value) != "foo-bar" {
			t.Errorf("expected message to be %s,but got %s", "foo-bar", string(event.Value))
		}
		return ProcessingSuccess
	})
	finalHandler := PipeHandlers(mw1, mw2)(actualHandler)
	finalHandler.HandleMessage(Message{}, context.Background())
}
