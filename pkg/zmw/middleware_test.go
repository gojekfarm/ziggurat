package zmw

import (
	"github.com/gojekfarm/ziggurat-go/pkg/mock"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zb"
	"github.com/gojekfarm/ziggurat-go/pkg/zlogger"
	"reflect"
	"testing"
	"time"
)

func TestMessageLogger_Success(t *testing.T) {
	handler := z.HandlerFunc(func(messageEvent zb.MessageEvent, app z.App) z.ProcessStatus {
		return z.ProcessingSuccess
	})
	ts := time.Time{}
	expectedArgs := map[string]interface{}{
		"ROUTE": "",
		"VALUE": "foo",
	}
	oldLogInfo := zlogger.LogInfo
	zlogger.LogInfo = func(msg string, args map[string]interface{}) {
		if !reflect.DeepEqual(args, expectedArgs) {
			t.Errorf("expected %v got %v", expectedArgs, args)
		}
	}
	defer func() {
		zlogger.LogInfo = oldLogInfo
	}()
	ml := MessageLogger(handler)
	event := zb.MessageEvent{
		MessageValueBytes: []byte("foo"),
		MessageKeyBytes:   []byte("foo"),
		Topic:             "",
		StreamRoute:       "",
		KafkaTimestamp:    ts,
		TimestampType:     "",
		Attributes:        nil,
	}

	ml.HandleMessage(event, mock.NewApp())

}
