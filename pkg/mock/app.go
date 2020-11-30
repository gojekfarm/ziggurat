package mock

import (
	"context"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zb"
)

type App struct {
	ContextFunc         func() context.Context
	RoutesFunc          func() []string
	MessageRetryFunc    func() z.MessageRetry
	MessageHandlerFunc  func() z.MessageHandler
	MetricPublisherFunc func() z.MetricPublisher
	HTTPServerFunc      func() z.Server
	ConfigStoreFunc     func() z.ConfigStore
}

func NewApp() *App {
	return &App{
		ContextFunc: func() context.Context {
			return context.Background()
		},
		RoutesFunc: func() []string {
			return []string{}
		},
		MessageHandlerFunc: func() z.MessageHandler {
			return z.HandlerFunc(func(messageEvent zb.MessageEvent, app z.App) z.ProcessStatus {
				return z.ProcessingSuccess
			})
		},
		HTTPServerFunc: func() z.Server {
			return nil
		},
		ConfigStoreFunc: func() z.ConfigStore {
			return NewConfigStore()
		},
		MessageRetryFunc: func() z.MessageRetry {
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

func (m *App) MessageRetry() z.MessageRetry {
	return m.MessageRetryFunc()
}

func (m *App) Handler() z.MessageHandler {
	return m.MessageHandlerFunc()
}

func (m *App) MetricPublisher() z.MetricPublisher {
	return m.MetricPublisher()
}

func (m *App) HTTPServer() z.Server {
	return m.HTTPServerFunc()
}

func (m *App) ConfigStore() z.ConfigStore {
	return m.ConfigStoreFunc()
}
