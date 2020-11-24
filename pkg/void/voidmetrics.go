package void

import (
	"fmt"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
)

type VoidMetrics struct{}

func NewVoidMetrics(store z.ConfigStore) z.MetricPublisher {
	return &VoidMetrics{}
}

func (v VoidMetrics) Start(app z.App) error {
	return fmt.Errorf("error starting metric plublisher, no implementation found")
}

func (v VoidMetrics) Stop() error {
	return nil
}

func (v VoidMetrics) IncCounter(metricName string, value int64, arguments map[string]string) error {
	return nil
}

func (v VoidMetrics) Gauge(metricName string, value int64, arguments map[string]string) error {
	return nil
}
