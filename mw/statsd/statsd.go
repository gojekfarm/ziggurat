package statsd

import (
	"context"
	"strconv"
	"time"

	"github.com/gojekfarm/ziggurat/logger"

	"github.com/cactus/go-statsd-client/v5/statsd"
	"github.com/gojekfarm/ziggurat"
)

type Client struct {
	client statsd.Statter
	host   string
	prefix string
	Logger ziggurat.StructuredLogger
}

const publishErrMsg = "statsd client: error publishing metric"

func NewPublisher(opts ...func(c *Client)) *Client {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}
	if c.prefix == "" {
		c.prefix = "ziggurat_statsd"
	}
	if c.host == "" {
		c.host = "localhost:8125"
	}

	if c.Logger == nil {
		c.Logger = logger.NewJSONLogger(logger.Disabled)
	}

	return c
}

func (s *Client) Run(ctx context.Context) error {
	config := &statsd.ClientConfig{
		Prefix:  s.prefix,
		Address: s.host,
	}
	client, clientErr := statsd.NewClientWithConfig(config)
	if clientErr != nil {
		return clientErr
	}
	s.client = client
	s.Logger.Info("starting go-routine publisher", map[string]interface{}{"publish-interval": "10s"})
	go func() {
		done := ctx.Done()
		<-done
		if s.client != nil {
			s.Logger.Error("error closing statsd client", s.client.Close())
		}
	}()
	go goRoutinePublisher(ctx, 10*time.Second, s)
	return nil
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

func (s *Client) PublishHandlerMetrics(handler ziggurat.Handler) ziggurat.Handler {
	return ziggurat.HandlerFunc(func(ctx context.Context, event ziggurat.Event) error {
		t1 := time.Now()
		err := handler.Handle(ctx, event)
		args := map[string]string{
			"route": event.Headers()[ziggurat.HeaderMessageRoute],
		}
		s.Logger.Error(publishErrMsg, s.Gauge("handler_execution_time", time.Since(t1).Milliseconds(), args))
		s.Logger.Error(publishErrMsg, s.IncCounter("message_count", 1, args))
		if err != nil {
			s.Logger.Error(publishErrMsg, s.IncCounter("processing_failure_count", 1, args))
			return err
		}
		s.Logger.Error(publishErrMsg, s.IncCounter("processing_success_count", 1, args))
		return err
	})
}

func (s *Client) PublishKafkaLag(handler ziggurat.Handler) ziggurat.Handler {
	return ziggurat.HandlerFunc(func(ctx context.Context, event ziggurat.Event) error {
		headers := event.Headers()
		topic := headers["x-kafka-topic"]
		partition := headers["x-kafka-partition"]
		timestampInt, parseErr := strconv.ParseInt(headers["x-kafka-timestamp"], 10, 64)
		if parseErr != nil {
			s.Logger.Error("error parsing timestamp", parseErr)
			return handler.Handle(ctx, event)
		}

		timestamp := time.Unix(timestampInt, 0)
		diff := time.Since(timestamp).Milliseconds()
		s.Logger.Error(publishErrMsg, s.Gauge("kafka_lag", diff, map[string]string{"topic": topic, "partition": partition}))
		return handler.Handle(ctx, event)

	})
}
