package zmw

import (
	"context"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zbasic"
	"github.com/gojekfarm/ziggurat-go/pkg/zlogger"
	"reflect"
	"testing"
	"time"
)

type mwMockApp struct{}

func (m mwMockApp) Context() context.Context {
	panic("implement me")
}

func (m mwMockApp) Routes() []string {
	panic("implement me")
}

func (m mwMockApp) MessageRetry() z.MessageRetry {
	panic("implement me")
}

func (m mwMockApp) Handler() z.MessageHandler {
	panic("implement me")
}

func (m mwMockApp) MetricPublisher() z.MetricPublisher {
	panic("implement me")
}

func (m mwMockApp) HTTPServer() z.Server {
	panic("implement me")
}

func (m mwMockApp) Config() *zbasic.Config {
	panic("implement me")
}

func (m mwMockApp) ConfigStore() z.ConfigStore {
	panic("implement me")
}

func TestMessageLogger_Success(t *testing.T) {
	handler := z.HandlerFunc(func(messageEvent zbasic.MessageEvent, app z.App) z.ProcessStatus {
		return z.ProcessingSuccess
	})
	ts := time.Time{}
	expectedArgs := map[string]interface{}{
		"ROUTE": "",
		"TOPIC": "",
		"K-TS":  ts.String(),
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
	event := zbasic.MessageEvent{
		MessageValueBytes: []byte("foo"),
		MessageKeyBytes:   []byte("foo"),
		Topic:             "",
		StreamRoute:       "",
		KafkaTimestamp:    ts,
		TimestampType:     "",
		Attributes:        nil,
	}

	ml.HandleMessage(event, mwMockApp{})

}
