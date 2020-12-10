package zvoid

import (
	"fmt"
	"github.com/gojekfarm/ziggurat/ztype"
)

type VoidMetrics struct{}

func NewMetrics() ztype.MetricPublisher {
	return &VoidMetrics{}
}

func (v VoidMetrics) Start(app ztype.App) error {
	return fmt.Errorf("error starting metric plublisher, no implementation found")
}

func (v VoidMetrics) Stop(a ztype.App) {

}

func (v VoidMetrics) IncCounter(metricName string, value int64, arguments map[string]string) error {
	return nil
}

func (v VoidMetrics) Gauge(metricName string, value int64, arguments map[string]string) error {
	return nil
}
