package zig

import (
	"testing"
)

func testMiddlewareFunc(next HandlerFunc) HandlerFunc {
	return func(messageEvent MessageEvent, app *App) ProcessStatus {
		messageEvent.MessageValueBytes = []byte("Test message")
		return next(messageEvent, app)
	}
}

func testMiddlewareAppender(next HandlerFunc) HandlerFunc {
	return func(messageEvent MessageEvent, app *App) ProcessStatus {
		msg := messageEvent.MessageValueBytes
		messageEvent.MessageValueBytes = append(msg, []byte(" appender")...)
		return next(messageEvent, app)
	}
}

func TestPipeHandlers(t *testing.T) {
	origHandler := func(msg MessageEvent, app *App) ProcessStatus {
		if string(msg.MessageValueBytes) == "Test message appender" {
			return ProcessingSuccess
		}
		return SkipMessage
	}
	handler := PipeHandlers(testMiddlewareFunc, testMiddlewareAppender)(origHandler)
	msgEvent := MessageEvent{}
	app := App{}

	if result := handler(msgEvent, &app); result != ProcessingSuccess {
		t.Errorf("Expected %v but got %v", ProcessingSuccess, result)
	}
}
