package metrics

import (
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"runtime"
	"strings"
	"time"
)

type StatsDConf struct {
	Host string `mapstructure:"host"`
}

type StatsD struct {
	client       statsd.Statter
	metricConfig *StatsDConf
	appName      string
}

func NewStatsD(config z.ConfigReader) z.MetricPublisher {
	metricConfig := parseStatsDConfig(config)
	return &StatsD{
		client:       nil,
		metricConfig: metricConfig,
		appName:      config.Config().ServiceName,
	}
}

func parseStatsDConfig(config z.ConfigReader) *StatsDConf {
	statsDConf := &StatsDConf{}
	if err := config.UnmarshalByKey("statsd", &statsDConf); err != nil {
		return &StatsDConf{Host: "localhost:8125"}
	}
	return statsDConf
}

func constructTags(tags map[string]string) string {
	var tagSlice []string
	for k, v := range tags {
		tagSlice = append(tagSlice, k+"="+v)
	}
	return strings.Join(tagSlice, ",")
}

func (s *StatsD) Start(app z.App) error {
	config := &statsd.ClientConfig{
		Prefix:  app.Config().ServiceName,
		Address: s.metricConfig.Host,
	}
	client, clientErr := statsd.NewClientWithConfig(config)
	if clientErr != nil {
		logger.LogError(clientErr, "ziggurat statsD", nil)
		return clientErr
	}
	s.client = client
	s.appName = app.Config().ServiceName
	go func() {
		logger.LogInfo("statsd: starting go-routine publisher", nil)
		done := app.Context().Done()
		t := time.NewTicker(10 * time.Second)
		tickerChan := t.C
		for {
			select {
			case <-done:
				t.Stop()
				logger.LogInfo("statsd: halting go-routine publisher", nil)
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
