package zr

import (
	"github.com/gojekfarm/ziggurat/pkg/mock"
	"github.com/gojekfarm/ziggurat/pkg/z"
	"github.com/gojekfarm/ziggurat/pkg/zb"
	"github.com/gojekfarm/ziggurat/pkg/zlog"
	"reflect"
	"testing"
	"time"
)

func TestDefaultRouter_HandleMessageError(t *testing.T) {
	called := false
	zlog.LogWarn = func(msg string, args map[string]interface{}) {
		called = true
	}
	dr := NewRouter()
	dr.HandleFunc("foo", func(event zb.MessageEvent, app z.App) z.ProcessStatus {
		return z.ProcessingSuccess
	})
	event := zb.MessageEvent{
		StreamRoute: "bar",
	}
	a := mock.NewApp()
	dr.HandleMessage(event, a)

	if !called {
		t.Errorf("expected warn logger to be called")
	}
}

func TestDefaultRouter_HandleMessage(t *testing.T) {
	dr := NewRouter()
	expectedEvent := zb.MessageEvent{
		MessageValueBytes: []byte("foo"),
		MessageKeyBytes:   []byte("foo"),
		Topic:             "baz",
		StreamRoute:       "foo",
		KafkaTimestamp:    time.Time{},
		TimestampType:     "",
		Attributes:        nil,
	}
	dr.HandleFunc("foo", func(event zb.MessageEvent, app z.App) z.ProcessStatus {
		if !reflect.DeepEqual(event, expectedEvent) {
			t.Errorf("expected event %+v, got %+v", expectedEvent, event)
		}
		return z.ProcessingSuccess
	})
	dr.HandleMessage(expectedEvent, mock.NewApp())
}

func TestDefaultRouter_Routes(t *testing.T) {
	dr := NewRouter()
	dr.HandleFunc("foo", func(event zb.MessageEvent, app z.App) z.ProcessStatus {
		return z.ProcessingSuccess
	})
	dr.HandleFunc("bar", func(event zb.MessageEvent, app z.App) z.ProcessStatus {
		return z.ProcessingSuccess
	})
	expectedRoutes := []string{"foo", "bar"}
	routes := dr.Routes()
	if !reflect.DeepEqual(expectedRoutes, routes) {
		t.Errorf("expected %v got %v", expectedRoutes, routes)
	}
}
