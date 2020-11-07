package zig

import (
	"testing"
)

func testMiddlewareFunc(next HandlerFunc) HandlerFunc {
	return func(messageEvent MessageEvent, app *Ziggurat) ProcessStatus {
		messageEvent.MessageValueBytes = []byte("Test message")
		return next(messageEvent, app)
	}
}

func testMiddlewareAppender(next HandlerFunc) HandlerFunc {
	return func(messageEvent MessageEvent, app *Ziggurat) ProcessStatus {
		msg := messageEvent.MessageValueBytes
		messageEvent.MessageValueBytes = append(msg, []byte(" appender")...)
		return next(messageEvent, app)
	}
}

func TestPipeHandlers(t *testing.T) {
	origHandler := func(msg MessageEvent, app *Ziggurat) ProcessStatus {
		if string(msg.MessageValueBytes) == "Test message appender" {
			return ProcessingSuccess
		}
		return SkipMessage
	}
	handler := pipeHandlers(testMiddlewareFunc, testMiddlewareAppender)(origHandler)
	msgEvent := MessageEvent{}
	app := Ziggurat{}

	if result := handler(msgEvent, &app); result != ProcessingSuccess {
		t.Errorf("Expected %v but got %v", ProcessingSuccess, result)
	}
}
