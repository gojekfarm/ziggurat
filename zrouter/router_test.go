package zrouter

import (
	"github.com/gojekfarm/ziggurat/mock"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/zlog"
	"github.com/gojekfarm/ziggurat/ztype"
	"reflect"
	"testing"
	"time"
)

func TestDefaultRouter_HandleMessageError(t *testing.T) {
	called := false
	zlog.LogWarn = func(msg string, args map[string]interface{}) {
		called = true
	}
	dr := New()
	dr.HandleFunc("foo", func(event zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
		return ztype.ProcessingSuccess
	})
	event := zbase.MessageEvent{
		StreamRoute: "bar",
	}
	a := mock.NewApp()
	dr.HandleMessage(event, a)

	if !called {
		t.Errorf("expected warn logger to be called")
	}
}

func TestDefaultRouter_HandleMessage(t *testing.T) {
	dr := New()
	expectedEvent := zbase.MessageEvent{
		MessageValueBytes: []byte("foo"),
		MessageKeyBytes:   []byte("foo"),
		Topic:             "baz",
		StreamRoute:       "foo",
		KafkaTimestamp:    time.Time{},
		TimestampType:     "",
		Attributes:        nil,
	}
	dr.HandleFunc("foo", func(event zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
		if !reflect.DeepEqual(event, expectedEvent) {
			t.Errorf("expected event %+v, got %+v", expectedEvent, event)
		}
		return ztype.ProcessingSuccess
	})
	dr.HandleMessage(expectedEvent, mock.NewApp())
}

func TestDefaultRouter_Routes(t *testing.T) {
	dr := New()
	dr.HandleFunc("foo", func(event zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
		return ztype.ProcessingSuccess
	})
	dr.HandleFunc("bar", func(event zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
		return ztype.ProcessingSuccess
	})
	expectedRoutes := []string{"foo", "bar"}
	routes := dr.Routes()
	if !reflect.DeepEqual(expectedRoutes, routes) {
		t.Errorf("expected %v got %v", expectedRoutes, routes)
	}
}
