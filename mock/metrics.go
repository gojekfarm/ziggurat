package mock

import (
	"github.com/gojekfarm/ziggurat/z"
)

type Metrics struct {
	StartFunc    func(a z.App) error
	StopFunc     func(a z.App)
	IncCountFunc func(metricName string, value int64, arguments map[string]string) error
	GaugeFunc    func(metricName string, value int64, arguments map[string]string) error
}

func NewMetrics() *Metrics {
	return &Metrics{
		StartFunc: func(a z.App) error {
			return nil
		},
		StopFunc: func(a z.App) {

		},
		IncCountFunc: func(metricName string, value int64, arguments map[string]string) error {
			return nil
		},
		GaugeFunc: func(metricName string, value int64, arguments map[string]string) error {
			return nil
		},
	}
}

func (m *Metrics) Stop(a z.App) {
	m.StopFunc(a)

}
func (m *Metrics) Start(app z.App) error {
	return m.StartFunc(app)
}

func (m *Metrics) IncCounter(metricName string, value int64, arguments map[string]string) error {
	return m.IncCountFunc(metricName, value, arguments)
}

func (m *Metrics) Gauge(metricName string, value int64, arguments map[string]string) error {
	return m.GaugeFunc(metricName, value, arguments)
}
