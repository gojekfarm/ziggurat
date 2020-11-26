package mock

import (
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
)

type MockRetry struct {
	StartFunc  func(app z.App) error
	RetryFunc  func(app z.App, payload basic.MessageEvent) error
	StopFunc   func() error
	ReplayFunc func() error
}

func NewMockRetry() *MockRetry {
	return &MockRetry{
		StartFunc: func(app z.App) error {
			return nil
		},
		RetryFunc: func(app z.App, payload basic.MessageEvent) error {
			return nil
		},
		StopFunc: func() error {
			return nil
		},
		ReplayFunc: func() error {
			return nil
		},
	}
}

func (m *MockRetry) Start(app z.App) error {
	return m.StartFunc(app)
}

func (m *MockRetry) Retry(app z.App, payload basic.MessageEvent) error {
	return m.Retry(app, payload)
}

func (m *MockRetry) Stop() error {
	return m.StopFunc()
}

func (m *MockRetry) Replay(app z.App, topicEntity string, count int) error {
	return m.ReplayFunc()
}
