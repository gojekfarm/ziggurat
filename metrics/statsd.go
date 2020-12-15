package metrics

import (
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/gojekfarm/ziggurat/zlog"
	"github.com/gojekfarm/ziggurat/ztype"
	"runtime"
	"strings"
	"time"
)

type StatsDClient struct {
	client statsd.Statter
	host   string
	prefix string
}

func NewStatsD(opts ...func(s *StatsDClient)) *StatsDClient {
	s := &StatsDClient{}
	for _, opt := range opts {
		opt(s)
	}
	if s.prefix == "" {
		s.prefix = "ziggurat_statsd"
	}
	if s.host == "" {
		s.host = "localhost:8125"
	}
	return s
}

func constructTags(tags map[string]string) string {
	var tagSlice []string
	for k, v := range tags {
		tagSlice = append(tagSlice, k+"="+v)
	}
	return strings.Join(tagSlice, ",")
}

func (s *StatsDClient) Start(app ztype.App) error {
	config := &statsd.ClientConfig{
		Prefix:  s.prefix,
		Address: s.host,
	}
	client, clientErr := statsd.NewClientWithConfig(config)
	if clientErr != nil {
		zlog.LogError(clientErr, "ziggurat statsD", nil)
		return clientErr
	}
	s.client = client
	go func() {
		zlog.LogInfo("statsd: starting go-routine publisher", nil)
		done := app.Context().Done()
		t := time.NewTicker(10 * time.Second)
		tickerChan := t.C
		for {
			select {
			case <-done:
				t.Stop()
				zlog.LogInfo("statsd: halting go-routine publisher", nil)
				return
			case <-tickerChan:
				s.client.Gauge("go_routine_count", int64(runtime.NumGoroutine()), 1.0)
			}
		}
	}()
	return nil
}

func (s *StatsDClient) Stop(a ztype.App) {
	if s.client != nil {
		zlog.LogError(s.client.Close(), "error stopping statsd client", nil)
	}
}

func (s *StatsDClient) constructFullMetricStr(metricName, tags string) string {
	return metricName + "," + tags + "," + "app_name=" + s.prefix
}

func (s *StatsDClient) IncCounter(metricName string, value int64, arguments map[string]string) error {
	tags := constructTags(arguments)
	finalMetricName := s.constructFullMetricStr(metricName, tags)

	return s.client.Inc(finalMetricName, value, 1.0)
}

func (s *StatsDClient) Gauge(metricName string, value int64, arguments map[string]string) error {
	tags := constructTags(arguments)
	finalMetricName := s.constructFullMetricStr(metricName, tags)
	return s.client.Gauge(finalMetricName, value, 1.0)
}
