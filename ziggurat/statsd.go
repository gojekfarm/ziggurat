package ziggurat

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

func setStatsDConfig(config Config, s *StatsD) {
	rawConfig := config.GetByKey("statsd")
	sanitizedConfig := rawConfig.(map[string]interface{})
	s.statsdConfig = &StatsDConf{host: sanitizedConfig["host"].(string)}
}

func constructTags(tags map[string]string) string {
	tagSlice := []string{}
	for k, v := range tags {
		tagSlice = append(tagSlice, k+"="+v)
	}
	return strings.Join(tagSlice, ",")
}

func (s *StatsD) Start(app *App) error {
	setStatsDConfig(*app.Config, s)
	config := &statsd.ClientConfig{
		Prefix:  app.Config.ServiceName,
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
