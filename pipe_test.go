package ziggurat

import (
	"testing"
)

func TestPipeHandlers(t *testing.T) {
	mw1 := func(next MessageHandler) MessageHandler {
		return HandlerFunc(func(messageEvent MessageEvent, z *Ziggurat) ProcessStatus {
			messageEvent.MessageValueBytes = []byte("foo")
			return next.HandleMessage(messageEvent, z)
		})
	}
	mw2 := func(next MessageHandler) MessageHandler {
		return HandlerFunc(func(messageEvent MessageEvent, z *Ziggurat) ProcessStatus {
			messageEvent.MessageValueBytes = append(messageEvent.MessageValueBytes, []byte("-bar")...)
			return next.HandleMessage(messageEvent, z)
		})
	}
	actualHandler := HandlerFunc(func(event MessageEvent, z *Ziggurat) ProcessStatus {
		if string(event.MessageValueBytes) != "foo-bar" {
			t.Errorf("expected message to be %s,but got %s", "foo-bar", string(event.MessageValueBytes))
		}
		return ProcessingSuccess
	})
	finalHandler := PipeHandlers(mw1, mw2)(actualHandler)
	finalHandler.HandleMessage(MessageEvent{}, NewApp())
}
