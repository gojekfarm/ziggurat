package ziggurat

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestDefaultRouter_HandleMessageError(t *testing.T) {
	called := false
	LogWarn = func(msg string, args map[string]interface{}) {
		called = true
	}
	dr := NewRouter()
	dr.HandleFunc("foo", func(event MessageEvent, app AppContext) ProcessStatus {
		return ProcessingSuccess
	})
	event := MessageEvent{
		StreamRoute: "bar",
	}
	a := NewZig()
	dr.HandleMessage(event, a)

	if !called {
		t.Errorf("expected warn logger to be called")
	}
}

func TestDefaultRouter_HandleMessage(t *testing.T) {
	dr := NewRouter()
	expectedEvent := MessageEvent{
		MessageValueBytes: []byte("foo"),
		MessageKeyBytes:   []byte("foo"),
		Topic:             "baz",
		StreamRoute:       "foo",
		ActualTimestamp:   time.Time{},
		TimestampType:     "",
		Attributes:        nil,
	}
	dr.HandleFunc("foo", func(event MessageEvent, app AppContext) ProcessStatus {
		if !reflect.DeepEqual(event, expectedEvent) {
			t.Errorf("expected event %+v, got %+v", expectedEvent, event)
		}
		return ProcessingSuccess
	})
	dr.HandleMessage(expectedEvent, NewZig())
}

func TestDefaultRouter_Routes(t *testing.T) {
	dr := NewRouter()
	dr.HandleFunc("foo", func(event MessageEvent, app AppContext) ProcessStatus {
		return ProcessingSuccess
	})
	dr.HandleFunc("bar", func(event MessageEvent, app AppContext) ProcessStatus {
		return ProcessingSuccess
	})
	expectedRoutes := sort.StringSlice([]string{"foo", "bar"})
	expectedRoutes.Sort()
	routes := sort.StringSlice(dr.Routes())
	routes.Sort()
	if !reflect.DeepEqual(expectedRoutes, routes) {
		t.Errorf("expected %v got %v", expectedRoutes, routes)
	}
}
