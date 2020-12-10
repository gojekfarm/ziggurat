package mock

import (
	"github.com/gojekfarm/ziggurat/ztype"
)

type Metrics struct {
	StartFunc    func(a ztype.App) error
	StopFunc     func(a ztype.App)
	IncCountFunc func(metricName string, value int64, arguments map[string]string) error
	GaugeFunc    func(metricName string, value int64, arguments map[string]string) error
}

func NewMetrics() *Metrics {
	return &Metrics{
		StartFunc: func(a ztype.App) error {
			return nil
		},
		StopFunc: func(a ztype.App) {

		},
		IncCountFunc: func(metricName string, value int64, arguments map[string]string) error {
			return nil
		},
		GaugeFunc: func(metricName string, value int64, arguments map[string]string) error {
			return nil
		},
	}
}

func (m *Metrics) Stop(a ztype.App) {
	m.StopFunc(a)

}
func (m *Metrics) Start(app ztype.App) error {
	return m.StartFunc(app)
}

func (m *Metrics) IncCounter(metricName string, value int64, arguments map[string]string) error {
	return m.IncCountFunc(metricName, value, arguments)
}

func (m *Metrics) Gauge(metricName string, value int64, arguments map[string]string) error {
	return m.GaugeFunc(metricName, value, arguments)
}
