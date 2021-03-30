package mw

import (
	"context"
	"errors"
	"github.com/gojekfarm/ziggurat/mock"
	"testing"

	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/logger"
)

func TestProcessingStatusLogger_LogStatus(t *testing.T) {
	expectedKVS := map[string]interface{}{"route": "foo", "value": []byte("bar")}
	dl := logger.DiscardLogger{
		InfoFunc: func(message string, kvs ...map[string]interface{}) {
			expectedMessage := "message processing succeeded"
			if message != expectedMessage {
				t.Errorf("expected message %s got %s", expectedMessage, message)
			}
			for k, _ := range kvs[0] {
				if _, ok := expectedKVS[k]; !ok {
					t.Errorf("expected key %s not found", k)
				}
			}
		},
		ErrorFunc: func(message string, err error, kvs ...map[string]interface{}) {
			expectedMessage := "message processing failed"
			for k, _ := range kvs[0] {
				if _, ok := expectedKVS[k]; !ok {
					t.Errorf("expected key %s not found", k)
				}
			}
			if message != expectedMessage {
				t.Errorf("expected message %s got %s", expectedMessage, message)
			}
			if err == nil {
				t.Errorf("expected error to be not nil")
			}
		},
	}

	sl := ProcessingStatusLogger{Logger: dl}

	me := mock.Event{
		ValueFunc: func() []byte {
			return []byte("bar")
		},
		HeadersFunc: func() map[string]string {
			return map[string]string{ziggurat.HeaderMessageRoute: "foo"}
		},
	}
	sl.LogStatus(ziggurat.HandlerFunc(func(ctx context.Context, event ziggurat.Event) error {
		return nil
	})).Handle(context.Background(), me)

	sl.LogStatus(ziggurat.HandlerFunc(func(ctx context.Context, event ziggurat.Event) error {
		return errors.New("error in handler")
	})).Handle(context.Background(), me)

}
