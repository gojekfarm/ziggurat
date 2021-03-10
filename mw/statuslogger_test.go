package mw

import (
	"context"
	"errors"
	"reflect"
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
			if !reflect.DeepEqual(kvs[0], expectedKVS) {
				t.Errorf("expected kvs %v got %v", expectedKVS, kvs)
			}
		},
		ErrorFunc: func(message string, err error, kvs ...map[string]interface{}) {
			expectedMessage := "message processing failed"
			if !reflect.DeepEqual(kvs[0], expectedKVS) {
				t.Errorf("expected kvs %v got %v", expectedKVS, kvs)
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

	me := ziggurat.MockEvent{
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
