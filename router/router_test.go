package router

import (
	"context"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/logger"
	"reflect"
	"testing"
)

func TestDefaultRouter_HandleMessageError(t *testing.T) {
	dr := New()
	dr.l = logger.NewJSONLogger("disabled")
	dr.HandleFunc("foo", func(event ziggurat.Event) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})
	event := ziggurat.CreateMessageEvent(nil, map[string]string{ziggurat.HeaderMessageRoute: "bar"}, context.Background())
	if status := dr.HandleEvent(event); status != ziggurat.SkipMessage {
		t.Errorf("expected status %d got status %d", ziggurat.SkipMessage, status)
	}

}

func TestDefaultRouter_HandleMessage(t *testing.T) {
	dr := New()
	expectedEvent := ziggurat.CreateMessageEvent(nil, map[string]string{ziggurat.HeaderMessageRoute: "foo"}, context.Background())
	dr.HandleFunc("foo", func(event ziggurat.Event) ziggurat.ProcessStatus {
		if !reflect.DeepEqual(event, expectedEvent) {
			t.Errorf("expected event %+v, got %+v", expectedEvent, event)
		}
		return ziggurat.ProcessingSuccess
	})
	dr.HandleEvent(ziggurat.CreateMessageEvent(nil, map[string]string{ziggurat.HeaderMessageRoute: "bar"}, context.Background()))
}
