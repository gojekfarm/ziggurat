package zmw

import (
	"github.com/gojekfarm/ziggurat/mock"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/zlog"
	"github.com/gojekfarm/ziggurat/ztype"
	"reflect"
	"testing"
	"time"
)

func TestMessageLogger_Success(t *testing.T) {
	handler := ztype.HandlerFunc(func(messageEvent zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
		return ztype.ProcessingSuccess
	})
	ts := time.Time{}
	expectedArgs := map[string]interface{}{
		"ROUTE": "",
		"VALUE": "foo",
	}
	oldLogInfo := zlog.LogInfo
	zlog.LogInfo = func(msg string, args map[string]interface{}) {
		if !reflect.DeepEqual(args, expectedArgs) {
			t.Errorf("expected %v got %v", expectedArgs, args)
		}
	}
	defer func() {
		zlog.LogInfo = oldLogInfo
	}()
	ml := MessageLogger(handler)
	event := zbase.MessageEvent{
		MessageValueBytes: []byte("foo"),
		MessageKeyBytes:   []byte("foo"),
		Topic:             "",
		StreamRoute:       "",
		ActualTimestamp:   ts,
		TimestampType:     "",
		Attributes:        nil,
	}

	ml.HandleMessage(event, mock.NewZig())

}
