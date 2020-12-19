package mw

import (
	"github.com/gojekfarm/ziggurat"
	"reflect"
	"testing"
	"time"
)

func TestMessageLogger_Success(t *testing.T) {
	handler := ziggurat.HandlerFunc(func(messageEvent ziggurat.Message, z *ziggurat.Ziggurat) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})
	ts := time.Time{}
	expectedArgs := map[string]interface{}{
		"ROUTE": "",
		"VALUE": "foo",
	}
	oldLogInfo := ziggurat.LogInfo
	ziggurat.LogInfo = func(msg string, args map[string]interface{}) {
		if !reflect.DeepEqual(args, expectedArgs) {
			t.Errorf("expected %v got %v", expectedArgs, args)
		}
	}
	defer func() {
		ziggurat.LogInfo = oldLogInfo
	}()
	ml := ProcessingStatusLogger(handler)
	event := ziggurat.Message{
		MessageValue:    []byte("foo"),
		MessageKey:      []byte("foo"),
		Topic:           "",
		StreamRoute:     "",
		ActualTimestamp: ts,
		TimestampType:   "",
		Attributes:      nil,
	}

	ml.HandleMessage(event, ziggurat.NewApp())

}
