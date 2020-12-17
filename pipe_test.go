package ziggurat

import (
	"testing"
	"time"
)

func TestPipeHandlers(t *testing.T) {
	mw1 := func(next MessageHandler) MessageHandler {
		return HandlerFunc(func(messageEvent Event, app AppContext) ProcessStatus {
			me := NewMessageEvent(nil, []byte("foo"), "", "", "", time.Time{})
			return next.HandleMessage(me, app)
		})
	}
	mw2 := func(next MessageHandler) MessageHandler {
		return HandlerFunc(func(messageEvent Event, app AppContext) ProcessStatus {
			newValue := append(messageEvent.MessageValue(), []byte("-bar")...)
			me := NewMessageEvent(nil, newValue, "", "", "", time.Time{})
			return next.HandleMessage(me, app)
		})
	}
	actualHandler := HandlerFunc(func(event Event, app AppContext) ProcessStatus {
		if string(event.MessageValue()) != "foo-bar" {
			t.Errorf("expected message to be %s,but got %s", "foo-bar", string(event.MessageValue()))
		}
		return ProcessingSuccess
	})
	finalHandler := PipeHandlers(mw1, mw2)(actualHandler)
	finalHandler.HandleMessage(&MessageEvent{}, NewZig())
}
