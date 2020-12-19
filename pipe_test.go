package ziggurat

import (
	"testing"
)

func TestPipeHandlers(t *testing.T) {
	mw1 := func(next Handler) Handler {
		return HandlerFunc(func(messageEvent Message, z *Ziggurat) ProcessStatus {
			messageEvent.MessageValue = []byte("foo")
			return next.HandleMessage(messageEvent, z)
		})
	}
	mw2 := func(next Handler) Handler {
		return HandlerFunc(func(messageEvent Message, z *Ziggurat) ProcessStatus {
			messageEvent.MessageValue = append(messageEvent.MessageValue, []byte("-bar")...)
			return next.HandleMessage(messageEvent, z)
		})
	}
	actualHandler := HandlerFunc(func(event Message, z *Ziggurat) ProcessStatus {
		if string(event.MessageValue) != "foo-bar" {
			t.Errorf("expected message to be %s,but got %s", "foo-bar", string(event.MessageValue))
		}
		return ProcessingSuccess
	})
	finalHandler := PipeHandlers(mw1, mw2)(actualHandler)
	finalHandler.HandleMessage(Message{}, NewApp())
}
