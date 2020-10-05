package ziggurat

import (
	"context"
)

type MetricPublisher interface {
	Start(ctx context.Context, app App) error
	Stop(ctx context.Context) error
	IncCounter(metricName string, value int, arguments map[string]string) error
}
