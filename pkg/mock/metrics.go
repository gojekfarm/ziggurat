package mock

import (
	"github.com/gojekfarm/ziggurat-go/pkg/z"
)

type Metrics struct {
	StartFunc    func(a z.App) error
	StopFunc     func() error
	IncCountFunc func(metricName string, value int64, arguments map[string]string) error
	GaugeFunc    func(metricName string, value int64, arguments map[string]string) error
}

func NewMockMetrics() *Metrics {
	return &Metrics{
		StartFunc: func(a z.App) error {
			return nil
		},
		StopFunc: func() error {
			return nil
		},
		IncCountFunc: func(metricName string, value int64, arguments map[string]string) error {
			return nil
		},
		GaugeFunc: func(metricName string, value int64, arguments map[string]string) error {
			return nil
		},
	}
}

func (m *Metrics) Start(app z.App) error {
	return m.StartFunc(app)
}

func (m *Metrics) Stop() error {
	return m.StopFunc()
}

func (m *Metrics) IncCounter(metricName string, value int64, arguments map[string]string) error {
	return m.IncCountFunc(metricName, value, arguments)
}

func (m *Metrics) Gauge(metricName string, value int64, arguments map[string]string) error {
	return m.GaugeFunc(metricName, value, arguments)
}
