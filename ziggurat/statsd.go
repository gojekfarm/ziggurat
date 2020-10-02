package ziggurat

import (
	"context"
	"github.com/cactus/go-statsd-client/statsd"
	"strings"
)

type StatsD struct {
	client statsd.Statter
}

func constructTags(tags map[string]string) string {
	tagSlice := []string{}
	for k, v := range tags {
		tagSlice = append(tagSlice, k+"="+v)
	}
	return strings.Join(tagSlice, ",")
}

func (s *StatsD) Start(ctx context.Context, applicationContext ApplicationContext) error {
	config := &statsd.ClientConfig{
		Prefix:  applicationContext.Config.ServiceName,
		Address: "127.0.0.1:8125",
	}
	client, clientErr := statsd.NewClientWithConfig(config)
	if clientErr != nil {
		MetricLogger.Error().Err(clientErr).Msg("")
		return clientErr
	}
	s.client = client
	return nil
}

func (s *StatsD) Stop(ctx context.Context) error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

func (s *StatsD) IncCounter(metricName string, value int, arguments map[string]string) error {
	tags := constructTags(arguments)
	finalMetricName := metricName + "," + tags
	return s.client.Inc(finalMetricName, int64(value), 1.0)
}
