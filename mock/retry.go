package mock

import (
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/ztype"
)

type Retry struct {
	StartFunc  func(app ztype.App) error
	RetryFunc  func(app ztype.App, payload zbase.MessageEvent) error
	StopFunc   func(app ztype.App)
	ReplayFunc func(app ztype.App, topicEntity string, count int) error
}

func NewRetry() *Retry {
	return &Retry{
		StartFunc: func(app ztype.App) error {
			return nil
		},
		RetryFunc: func(app ztype.App, payload zbase.MessageEvent) error {
			return nil
		},
		StopFunc: func(a ztype.App) {

		},
		ReplayFunc: func(app ztype.App, topicEntity string, count int) error {
			return nil
		},
	}
}

func (m *Retry) Start(app ztype.App) error {
	return m.StartFunc(app)
}

func (m *Retry) Retry(app ztype.App, payload zbase.MessageEvent) error {
	return m.RetryFunc(app, payload)
}

func (m *Retry) Stop(app ztype.App) {
	m.StopFunc(app)
}

func (m *Retry) Replay(app ztype.App, topicEntity string, count int) error {
	return m.ReplayFunc(app, topicEntity, count)
}
