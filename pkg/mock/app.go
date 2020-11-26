package mock

import (
	"context"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
)

type App struct {
	ContextFunc         func() context.Context
	RoutesFunc          func() []string
	MessageRetryFunc    func() z.MessageRetry
	MessageHandlerFunc  func() z.MessageHandler
	MetricPublisherFunc func() z.MetricPublisher
	HTTPServerFunc      func() z.HttpServer
	ConfigFunc          func() *basic.Config
	ConfigStoreFunc     func() z.ConfigStore
}

func NewMockApp() *App {
	return &App{
		ContextFunc: func() context.Context {
			return context.Background()
		},
		RoutesFunc: func() []string {
			return []string{}
		},
		MessageHandlerFunc: func() z.MessageHandler {
			return z.HandlerFunc(func(messageEvent basic.MessageEvent, app z.App) z.ProcessStatus {
				return z.ProcessingSuccess
			})
		},
		HTTPServerFunc: func() z.HttpServer {
			return nil
		},
		ConfigFunc: func() *basic.Config {
			return &basic.Config{}
		},
		ConfigStoreFunc: func() z.ConfigStore {
			return nil
		},
	}
}

func (m *App) Context() context.Context {
	return m.ContextFunc()
}

func (m *App) Routes() []string {
	return m.RoutesFunc()
}

func (m *App) MessageRetry() z.MessageRetry {
	return m.MessageRetryFunc()
}

func (m *App) Handler() z.MessageHandler {
	return m.MessageHandlerFunc()
}

func (m *App) MetricPublisher() z.MetricPublisher {
	return m.MetricPublisher()
}

func (m *App) HTTPServer() z.HttpServer {
	return m.HTTPServerFunc()
}

func (m *App) Config() *basic.Config {
	return m.ConfigFunc()
}

func (m *App) ConfigStore() z.ConfigStore {
	return m.ConfigStore()
}
