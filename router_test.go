package ziggurat

import (
	"reflect"
	"testing"
)

func TestDefaultRouter_HandleMessageError(t *testing.T) {
	dr := NewRouter()
	dr.l = NewLogger("disabled")
	dr.HandleFunc("foo", func(event Event) ProcessStatus {
		return ProcessingSuccess
	})
	event := Message{
		MessageHeaders: map[string]string{HeaderMessageRoute: "bar"},
	}
	if status := dr.HandleEvent(event); status != SkipMessage {
		t.Errorf("expected status %d got status %d", SkipMessage, status)
	}

}

func TestDefaultRouter_HandleMessage(t *testing.T) {
	dr := NewRouter()
	expectedEvent := Message{
		MessageHeaders: map[string]string{HeaderMessageType: "kafka", HeaderMessageRoute: "foo"},
	}
	dr.HandleFunc("foo", func(event Event) ProcessStatus {
		if !reflect.DeepEqual(event, expectedEvent) {
			t.Errorf("expected event %+v, got %+v", expectedEvent, event)
		}
		return ProcessingSuccess
	})
	dr.HandleEvent(Message{
		MessageHeaders: map[string]string{HeaderMessageRoute: "foo", HeaderMessageType: "kafka"},
	})
}
