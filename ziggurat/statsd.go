package ziggurat

import (
	"context"
	"github.com/alexcesaro/statsd"
	"runtime"
)

type Statsd struct {
	client *statsd.Client
}

func (s *Statsd) Start(ctx context.Context, applicationContext ApplicationContext) error {
	client, openErr := statsd.New()
	logErr(openErr, "statsd error")
	s.client = client
	return openErr
}

func (s *Statsd) PublishMetric(applicationContext ApplicationContext, metricName string, args map[string]interface{}) error {
	s.client.Gauge("num_goroutine", runtime.NumGoroutine())
	return nil
}

func (s *Statsd) Stop(ctx context.Context) error {
	if s.client != nil {
		s.client.Close()
	}
	return nil
}
