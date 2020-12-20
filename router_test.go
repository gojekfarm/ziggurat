package ziggurat

import (
	"context"
	"reflect"
	"testing"
)

func TestDefaultRouter_HandleMessageError(t *testing.T) {
	dr := NewRouter()
	dr.l = NewLogger("disabled")
	dr.HandleFunc("foo", func(event Message, ctx context.Context) ProcessStatus {
		return ProcessingSuccess
	})
	event := Message{
		RouteName: "bar",
	}
	if status := dr.HandleMessage(event, context.Background()); status != SkipMessage {
		t.Errorf("expected status %d got status %d", SkipMessage, status)
	}

}

func TestDefaultRouter_HandleMessage(t *testing.T) {
	dr := NewRouter()
	expectedEvent := Message{
		RouteName:  "foo",
		Attributes: MsgAttributes{"bar": "baz"},
	}
	dr.HandleFunc("foo", func(event Message, ctx context.Context) ProcessStatus {
		if !reflect.DeepEqual(event, expectedEvent) {
			t.Errorf("expected event %+v, got %+v", expectedEvent, event)
		}
		return ProcessingSuccess
	})
	dr.HandleMessage(Message{RouteName: "foo", Attributes: MsgAttributes{"bar": "baz"}}, context.Background())
}
