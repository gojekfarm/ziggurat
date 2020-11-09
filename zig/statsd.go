package zig

import (
	"github.com/cactus/go-statsd-client/statsd"
	"runtime"
	"strings"
	"time"
)

type MetricConfig struct {
	host string
}

type StatsD struct {
	client       statsd.Statter
	metricConfig *MetricConfig
	appName      string
}

func NewStatsD(config ConfigReader) MetricPublisher {
	metricConfig := parseStatsDConfig(config)
	return &StatsD{
		client:       nil,
		metricConfig: metricConfig,
		appName:      config.Config().ServiceName,
	}
}

func parseStatsDConfig(config ConfigReader) *MetricConfig {
	rawConfig := config.GetByKey("statsd")
	if sanitizedConfig, ok := rawConfig.(map[string]interface{}); !ok {
		metricLogger.Error().Err(ErrParsingStatsDConfig).Msg("[ZIG STATSD]")
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

func (s *StatsD) Start(app App) error {
	config := &statsd.ClientConfig{
		Prefix:  app.Config().ServiceName,
		Address: s.metricConfig.host,
	}
	client, clientErr := statsd.NewClientWithConfig(config)
	if clientErr != nil {
		metricLogger.Error().Err(clientErr).Msg("[ZIG STATSD]")
		return clientErr
	}
	s.client = client
	s.appName = app.Config().ServiceName
	go func() {
		logInfo("statsd: starting go-routine publisher")
		done := app.Context().Done()
		t := time.NewTicker(10 * time.Second)
		tickerChan := t.C
		for {
			select {
			case <-done:
				t.Stop()
				logInfo("halting go-routine publisher")
				return
			case <-tickerChan:
				s.client.Gauge("go_routine_count", int64(runtime.NumGoroutine()), 1.0)
			}
		}
	}()
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
