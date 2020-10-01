package ziggurat

import (
	"context"
	"github.com/cactus/go-statsd-client/statsd"
	"runtime"
)

type StatsD struct {
	client statsd.Statter
}

func (s *StatsD) Start(ctx context.Context, applicationContext ApplicationContext) error {
	config := &statsd.ClientConfig{
		Address: "127.0.0.1:8125",
		Prefix:  applicationContext.Config.ServiceName,
	}
	client, clientErr := statsd.NewClientWithConfig(config)
	if clientErr != nil {
		return clientErr
	}
	s.client = client
	return nil

}

func (s *StatsD) PublishMetric(applicationContext ApplicationContext, metricName string, args map[string]interface{}) error {
	sendErr := s.client.Gauge(metricName, int64(runtime.NumGoroutine()), 1.0)
	return sendErr
}

func (s *StatsD) Stop(ctx context.Context) error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}
