package mock

import (
	"context"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/ztype"
)

type App struct {
	ContextFunc         func() context.Context
	RoutesFunc          func() []string
	MessageRetryFunc    func() ztype.MessageRetry
	MessageHandlerFunc  func() ztype.MessageHandler
	MetricPublisherFunc func() ztype.MetricPublisher
	HTTPServerFunc      func() ztype.Server
	ConfigStoreFunc     func() ztype.ConfigStore
}

func NewApp() *App {
	return &App{
		ContextFunc: func() context.Context {
			return context.Background()
		},
		RoutesFunc: func() []string {
			return []string{}
		},
		MessageHandlerFunc: func() ztype.MessageHandler {
			return ztype.HandlerFunc(func(messageEvent zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
				return ztype.ProcessingSuccess
			})
		},
		HTTPServerFunc: func() ztype.Server {
			return nil
		},
		ConfigStoreFunc: func() ztype.ConfigStore {
			return NewConfigStore()
		},
		MessageRetryFunc: func() ztype.MessageRetry {
			return NewRetry()
		},
	}
}

func (m *App) Context() context.Context {
	return m.ContextFunc()
}

func (m *App) Routes() []string {
	return m.RoutesFunc()
}

func (m *App) MessageRetry() ztype.MessageRetry {
	return m.MessageRetryFunc()
}

func (m *App) Handler() ztype.MessageHandler {
	return m.MessageHandlerFunc()
}

func (m *App) MetricPublisher() ztype.MetricPublisher {
	return m.MetricPublisher()
}

func (m *App) HTTPServer() ztype.Server {
	return m.HTTPServerFunc()
}

func (m *App) ConfigStore() ztype.ConfigStore {
	return m.ConfigStoreFunc()
}
