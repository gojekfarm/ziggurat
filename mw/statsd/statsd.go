package statsd

import (
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/gojekfarm/ziggurat"
	"time"
)

type Client struct {
	client  statsd.Statter
	host    string
	prefix  string
	handler ziggurat.MessageHandler
}

func NewClient(opts ...func(s *Client)) *Client {
	s := &Client{}
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

func (s *Client) Start(z *ziggurat.Ziggurat) error {
	config := &statsd.ClientConfig{
		Prefix:  s.prefix,
		Address: s.host,
	}
	client, clientErr := statsd.NewClientWithConfig(config)
	if clientErr != nil {
		ziggurat.LogError(clientErr, "ziggurat statsD", nil)
		return clientErr
	}
	s.client = client
	go GoRoutinePublisher(z.Context(), 10*time.Second, s)
	return nil
}

func (s *Client) Stop() {
	if s.client != nil {
		ziggurat.LogError(s.client.Close(), "error stopping statsd client", nil)
	}
}

func (s *Client) constructFullMetricStr(metricName, tags string) string {
	return metricName + "," + tags + "," + "app_name=" + s.prefix
}

func (s *Client) IncCounter(metricName string, value int64, arguments map[string]string) error {
	tags := constructTags(arguments)
	finalMetricName := s.constructFullMetricStr(metricName, tags)

	return s.client.Inc(finalMetricName, value, 1.0)
}

func (s *Client) Gauge(metricName string, value int64, arguments map[string]string) error {
	tags := constructTags(arguments)
	finalMetricName := s.constructFullMetricStr(metricName, tags)
	return s.client.Gauge(finalMetricName, value, 1.0)
}

func (s *Client) PublishHandlerMetrics(handler ziggurat.MessageHandler) ziggurat.MessageHandler {
	return ziggurat.HandlerFunc(func(messageEvent ziggurat.MessageEvent, z *ziggurat.Ziggurat) ziggurat.ProcessStatus {
		arguments := map[string]string{"route": messageEvent.StreamRoute}
		startTime := time.Now()
		status := handler.HandleMessage(messageEvent, z)
		endTime := time.Now()
		diffTimeInMS := endTime.Sub(startTime).Milliseconds()
		s.Gauge("handler_func_exec_time", diffTimeInMS, arguments)
		switch status {
		case ziggurat.RetryMessage, ziggurat.SkipMessage:
			s.IncCounter("message_processing_failure_skip_count", 1, arguments)
		default:
			s.IncCounter("message_processing_success_count", 1, arguments)
		}
		return status
	})
}

func (s *Client) PublishKafkaLag(handler ziggurat.MessageHandler) ziggurat.MessageHandler {
	return ziggurat.HandlerFunc(func(messageEvent ziggurat.MessageEvent, z *ziggurat.Ziggurat) ziggurat.ProcessStatus {
		actualTS := messageEvent.ActualTimestamp
		now := time.Now()
		diff := now.Sub(actualTS).Milliseconds()
		s.Gauge("kafka_message_lag", diff, map[string]string{
			"route": messageEvent.StreamRoute,
		})
		return handler.HandleMessage(messageEvent, z)
	})
}
