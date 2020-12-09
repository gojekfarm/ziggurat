package void

import (
	"fmt"
	"github.com/gojekfarm/ziggurat/z"
)

type VoidMetrics struct{}

func NewMetrics(store z.ConfigStore) z.MetricPublisher {
	return &VoidMetrics{}
}

func (v VoidMetrics) Start(app z.App) error {
	return fmt.Errorf("error starting metric plublisher, no implementation found")
}

func (v VoidMetrics) Stop(a z.App) {

}

func (v VoidMetrics) IncCounter(metricName string, value int64, arguments map[string]string) error {
	return nil
}

func (v VoidMetrics) Gauge(metricName string, value int64, arguments map[string]string) error {
	return nil
}
