package ziggurat

import "context"

type MetricPublisher interface {
	Start(ctx context.Context, applicationContext ApplicationContext) error
	PublishMetric(applicationContext ApplicationContext, metricName string, arguments map[string]interface{}) error
	Stop(ctx context.Context) error
}
