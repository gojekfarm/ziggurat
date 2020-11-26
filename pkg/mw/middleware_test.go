package mw

import (
	"context"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
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

func (m mwMockApp) Config() *basic.Config {
	panic("implement me")
}

func (m mwMockApp) ConfigStore() z.ConfigStore {
	panic("implement me")
}

func TestMessageLogger_Success(t *testing.T) {
	handler := z.HandlerFunc(func(messageEvent basic.MessageEvent, app z.App) z.ProcessStatus {
		return z.ProcessingSuccess
	})
	ts := time.Time{}
	expectedArgs := map[string]interface{}{
		"ROUTE": "",
		"TOPIC": "",
		"K-TS":  ts.String(),
		"VALUE": "foo",
	}
	oldLogInfo := logger.LogInfo
	logger.LogInfo = func(msg string, args map[string]interface{}) {
		if !reflect.DeepEqual(args, expectedArgs) {
			t.Errorf("expected %v got %v", expectedArgs, args)
		}
	}
	defer func() {
		logger.LogInfo = oldLogInfo
	}()
	ml := MessageLogger(handler)
	event := basic.MessageEvent{
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
