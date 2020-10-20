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

func (s *StatsD) Start(app *App) error {
	parsedConfig := parseStatsDConfig(app.config)
	s.statsdConfig = parsedConfig
	config := &statsd.ClientConfig{
		Prefix:  app.config.ServiceName,
		Address: s.statsdConfig.host,
	}
	client, clientErr := statsd.NewClientWithConfig(config)
	if clientErr != nil {
		metricLogger.Error().Err(clientErr).Msg("")
		return clientErr
	}
	s.client = client
	return nil
}

func (s *StatsD) Stop() error {
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
