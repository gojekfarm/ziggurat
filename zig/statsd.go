package zig

import (
	"github.com/cactus/go-statsd-client/statsd"
	"strings"
)

type MetricConfig struct {
	host string
}

type StatsD struct {
	client       statsd.Statter
	metricConfig *MetricConfig
	appName      string
}

func NewStatsD(config *Config) MetricPublisher {
	metricConfig := parseStatsDConfig(config)
	return &StatsD{
		client:       nil,
		metricConfig: metricConfig,
		appName:      config.ServiceName,
	}
}

func parseStatsDConfig(config *Config) *MetricConfig {
	rawConfig := config.GetByKey("statsd")
	if sanitizedConfig, ok := rawConfig.(map[string]interface{}); !ok {
		metricLogger.Error().Err(ErrParsingStatsDConfig).Msg("")
		return &MetricConfig{host: "localhost:8125"}
	} else {
		return &MetricConfig{host: sanitizedConfig["host"].(string)}
	}
}

func constructTags(tags map[string]string) string {
	var tagSlice []string
	for k, v := range tags {
		tagSlice = append(tagSlice, k+"="+v)
	}
	return strings.Join(tagSlice, ",")
}

func (s *StatsD) Start(app *App) error {
	config := &statsd.ClientConfig{
		Prefix:  app.config.ServiceName,
		Address: s.metricConfig.host,
	}
	client, clientErr := statsd.NewClientWithConfig(config)
	if clientErr != nil {
		metricLogger.Error().Err(clientErr).Msg("")
		return clientErr
	}
	s.client = client
	s.appName = app.config.ServiceName
	return nil
}

func (s *StatsD) Stop() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

func (s *StatsD) constructFullMetricStr(metricName, tags string) string {
	return metricName + "," + tags + "," + "app_name=" + s.appName
}

func (s *StatsD) IncCounter(metricName string, value int64, arguments map[string]string) error {
	tags := constructTags(arguments)
	finalMetricName := s.constructFullMetricStr(metricName, tags)
	return s.client.Inc(finalMetricName, value, 1.0)
}

func (s *StatsD) Gauge(metricName string, value int64, arguments map[string]string) error {
	tags := constructTags(arguments)
	finalMetricName := s.constructFullMetricStr(metricName, tags)
	return s.client.Gauge(finalMetricName, value, 1.0)
}
