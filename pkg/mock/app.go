package mock

import (
	"context"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
)

type MockApp struct {
	ContextFunc         func() context.Context
	RoutesFunc          func() []string
	MessageRetryFunc    func() z.MessageRetry
	MessageHandlerFunc  func() z.MessageHandler
	MetricPublisherFunc func() z.MetricPublisher
	HTTPServerFunc      func() z.HttpServer
	ConfigFunc          func() *basic.Config
	ConfigStoreFunc     func() z.ConfigStore
}

func (m *MockApp) Context() context.Context {
	return m.ContextFunc()
}

func (m *MockApp) Routes() []string {
	return m.RoutesFunc()
}

func (m *MockApp) MessageRetry() z.MessageRetry {
	return m.MessageRetryFunc()
}

func (m *MockApp) Handler() z.MessageHandler {
	return m.MessageHandlerFunc()
}

func (m *MockApp) MetricPublisher() z.MetricPublisher {
	return m.MetricPublisher()
}

func (m *MockApp) HTTPServer() z.HttpServer {
	return m.HTTPServerFunc()
}

func (m *MockApp) Config() *basic.Config {
	return m.ConfigFunc()
}

func (m *MockApp) ConfigStore() z.ConfigStore {
	return m.ConfigStore()
}
