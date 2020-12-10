package zrouter

import (
	"github.com/gojekfarm/ziggurat/mock"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/ztype"
	"testing"
)

func TestPipeHandlers(t *testing.T) {
	mw1 := func(next ztype.MessageHandler) ztype.MessageHandler {
		return ztype.HandlerFunc(func(messageEvent zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
			messageEvent.MessageValueBytes = []byte("foo")
			return next.HandleMessage(messageEvent, app)
		})
	}
	mw2 := func(next ztype.MessageHandler) ztype.MessageHandler {
		return ztype.HandlerFunc(func(messageEvent zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
			messageEvent.MessageValueBytes = append(messageEvent.MessageValueBytes, []byte("-bar")...)
			return next.HandleMessage(messageEvent, app)
		})
	}
	actualHandler := ztype.HandlerFunc(func(event zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
		if string(event.MessageValueBytes) != "foo-bar" {
			t.Errorf("expected message to be %s,but got %s", "foo-bar", string(event.MessageValueBytes))
		}
		return ztype.ProcessingSuccess
	})
	finalHandler := PipeHandlers(mw1, mw2)(actualHandler)
	finalHandler.HandleMessage(zbase.MessageEvent{}, mock.NewApp())
}
