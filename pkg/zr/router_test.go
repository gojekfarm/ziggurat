package zr

import (
	"github.com/gojekfarm/ziggurat-go/pkg/mock"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zb"
	"github.com/gojekfarm/ziggurat-go/pkg/zlog"
	"reflect"
	"testing"
	"time"
)

func TestDefaultRouter_HandleMessageError(t *testing.T) {
	oldLogFatal := zlog.LogFatal
	called := false
	zlog.LogWarn = func(msg string, args map[string]interface{}) {
		called = true
	}
	defer func() {
		zlog.LogFatal = oldLogFatal
	}()
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
