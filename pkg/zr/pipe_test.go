package zr

import (
	"github.com/gojekfarm/ziggurat-go/pkg/mock"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zbasic"
	"testing"
)

func TestPipeHandlers(t *testing.T) {
	mw1 := func(next z.MessageHandler) z.MessageHandler {
		return z.HandlerFunc(func(messageEvent zbasic.MessageEvent, app z.App) z.ProcessStatus {
			messageEvent.MessageValueBytes = []byte("foo")
			return next.HandleMessage(messageEvent, app)
		})
	}
	mw2 := func(next z.MessageHandler) z.MessageHandler {
		return z.HandlerFunc(func(messageEvent zbasic.MessageEvent, app z.App) z.ProcessStatus {
			messageEvent.MessageValueBytes = append(messageEvent.MessageValueBytes, []byte("-bar")...)
			return next.HandleMessage(messageEvent, app)
		})
	}
	actualHandler := z.HandlerFunc(func(event zbasic.MessageEvent, app z.App) z.ProcessStatus {
		if string(event.MessageValueBytes) != "foo-bar" {
			t.Errorf("expected message to be %s,but got %s", "foo-bar", string(event.MessageValueBytes))
		}
		return z.ProcessingSuccess
	})
	finalHandler := PipeHandlers(mw1, mw2)(actualHandler)
	finalHandler.HandleMessage(zbasic.MessageEvent{}, mock.NewApp())
}
