package mock

import (
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zb"
)

type Retry struct {
	StartFunc  func(app z.App) error
	RetryFunc  func(app z.App, payload zb.MessageEvent) error
	StopFunc   func(app z.App)
	ReplayFunc func(app z.App, topicEntity string, count int) error
}

func NewRetry() *Retry {
	return &Retry{
		StartFunc: func(app z.App) error {
			return nil
		},
		RetryFunc: func(app z.App, payload zb.MessageEvent) error {
			return nil
		},
		StopFunc: func(a z.App) {

		},
		ReplayFunc: func(app z.App, topicEntity string, count int) error {
			return nil
		},
	}
}

func (m *Retry) Start(app z.App) error {
	return m.StartFunc(app)
}

func (m *Retry) Retry(app z.App, payload zb.MessageEvent) error {
	return m.RetryFunc(app, payload)
}

func (m *Retry) Stop(app z.App) {
	m.StopFunc(app)
}

func (m *Retry) Replay(app z.App, topicEntity string, count int) error {
	return m.ReplayFunc(app, topicEntity, count)
}
