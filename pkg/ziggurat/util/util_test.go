package util

import (
	"github.com/gojekfarm/ziggurat-go/pkg/ziggurat/basic"
	at "github.com/gojekfarm/ziggurat-go/pkg/ziggurat/z"
	"github.com/gojekfarm/ziggurat-go/pkg/ziggurat/zig"
	"testing"
)

func testMiddlewareFunc(next at.HandlerFunc) at.HandlerFunc {
	return func(messageEvent basic.MessageEvent, app at.App) at.ProcessStatus {
		messageEvent.MessageValueBytes = []byte("Test message")
		return next(messageEvent, app)
	}
}

func testMiddlewareAppender(next at.HandlerFunc) at.HandlerFunc {
	return func(messageEvent basic.MessageEvent, app at.App) at.ProcessStatus {
		msg := messageEvent.MessageValueBytes
		messageEvent.MessageValueBytes = append(msg, []byte(" appender")...)
		return next(messageEvent, app)
	}
}

func TestPipeHandlers(t *testing.T) {
	origHandler := func(msg basic.MessageEvent, app at.App) at.ProcessStatus {
		if string(msg.MessageValueBytes) == "Test message appender" {
			return at.ProcessingSuccess
		}
		return at.SkipMessage
	}
	handler := PipeHandlers(testMiddlewareFunc, testMiddlewareAppender)(origHandler)
	msgEvent := basic.MessageEvent{}
	app := zig.NewApp()

	if result := handler(msgEvent, app); result != at.ProcessingSuccess {
		t.Errorf("Expected %v but got %v", at.ProcessingSuccess, result)
	}
}
