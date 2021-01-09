package router

import (
	"context"
	"github.com/gojekfarm/ziggurat"
	"reflect"
	"testing"
)

func TestDefaultRouter_HandleMessage(t *testing.T) {
	dr := New()
	expectedEvent := ziggurat.CreateMockEvent()
	expectedEvent.ValueFunc = func() []byte {
		return nil
	}
	expectedEvent.HeadersFunc = func() map[string]string {
		return map[string]string{ziggurat.HeaderMessageRoute: "bar"}
	}
	dr.HandleFunc("foo", func(event ziggurat.Event) ziggurat.ProcessStatus {
		if !reflect.DeepEqual(event, expectedEvent) {
			t.Errorf("expected event %+v, got %+v", expectedEvent, event)
		}
		return ziggurat.ProcessingSuccess
	})
	dr.HandleEvent(ziggurat.MockEvent{
		ValueFunc: func() []byte {
			return nil
		},
		HeadersFunc: func() map[string]string {
			return map[string]string{ziggurat.HeaderMessageRoute: "bar"}
		},
		ContextFunc: func() context.Context {
			return context.Background()
		},
	})
}

func TestDefaultRouter_NotFoundHandler(t *testing.T) {
	notFoundHandlerCalled := false
	dr := New(WithNotFoundHandler(func(event ziggurat.Event) ziggurat.ProcessStatus {
		notFoundHandlerCalled = true
		return ziggurat.ProcessingSuccess
	}))

	dr.HandleFunc("foo", func(event ziggurat.Event) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})

	dr.HandleEvent(ziggurat.MockEvent{
		ValueFunc: func() []byte {
			return []byte{}
		},
		HeadersFunc: func() map[string]string {
			return map[string]string{ziggurat.HeaderMessageRoute: "bar"}
		},
		ContextFunc: func() context.Context {
			return context.Background()
		},
	})

	if !notFoundHandlerCalled {
		t.Errorf("expected %v got %v", true, notFoundHandlerCalled)
	}

}
