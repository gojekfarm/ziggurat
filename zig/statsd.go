package zig

import (
	"github.com/cactus/go-statsd-client/statsd"
	"strings"
)

type StatsDConf struct {
	host string
}

type StatsD struct {
	client       statsd.Statter
	statsdConfig *StatsDConf
	appName      string
}

func NewStatsD(app *App) MetricPublisher {
	cfg := parseStatsDConfig(app.config)
	s := &StatsD{
		statsdConfig: cfg,
		appName:      app.config.ServiceName,
	}
	return s
}

func parseStatsDConfig(config *Config) *StatsDConf {
	rawConfig := config.GetByKey("statsd")
	if sanitizedConfig, ok := rawConfig.(map[string]interface{}); !ok {
		metricLogger.Error().Err(ErrParsingStatsDConfig).Msg("")
		return &StatsDConf{host: "localhost:8125"}
	} else {
		return &StatsDConf{host: sanitizedConfig["host"].(string)}
	}
}

func constructTags(tags map[string]string) string {
	var tagSlice []string
	for k, v := range tags {
		tagSlice = append(tagSlice, k+"="+v)
	}
	return strings.Join(tagSlice, ",")
}

func (s *StatsD) Start(app *App) (chan int, error) {
	stopNotifierChan := make(chan int)
	clientConf := &statsd.ClientConfig{
		Address: s.statsdConfig.host,
		Prefix:  s.appName,
	}
	if client, err := statsd.NewClientWithConfig(clientConf); err != nil {
		go func() {
			close(stopNotifierChan)
		}()
		return stopNotifierChan, err
	} else {
		s.client = client
		return stopNotifierChan, nil
	}
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
