package statsd

import (
	"context"
	"strconv"
	"time"

	"github.com/gojekfarm/ziggurat/kafka"

	"github.com/cactus/go-statsd-client/v5/statsd"
	"github.com/gojekfarm/ziggurat"
)

type Client struct {
	client  statsd.Statter
	host    string
	prefix  string
	handler ziggurat.Handler
}

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
	go func() {
		done := ctx.Done()
		<-done
		if s.client != nil {
			s.client.Close()
		}
	}()
	go GoRoutinePublisher(ctx, 10*time.Second, s)
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
		t2 := time.Now()
		args := map[string]string{
			"route": event.Headers()[ziggurat.HeaderMessageRoute],
		}
		s.Gauge("handler_execution_time", t2.Sub(t1).Milliseconds(), args)
		if err != nil {
			s.IncCounter("processing_failure_count", 1, args)
			return err
		}
		s.IncCounter("processing_success_count", 1, args)
		return err
	})
}

func (s *Client) PublishKafkaLag(handler ziggurat.Handler) ziggurat.Handler {
	return ziggurat.HandlerFunc(func(ctx context.Context, event ziggurat.Event) error {
		if kafkaMsg, ok := event.(kafka.Message); !ok {
			return handler.Handle(ctx, event)
		} else {
			diff := time.Now().Sub(kafkaMsg.Timestamp).Milliseconds()
			s.Gauge("kafka_lag", diff, map[string]string{"topic": kafkaMsg.Topic, "partition": strconv.Itoa(kafkaMsg.Partition)})
		}
		return handler.Handle(ctx, event)
	})
}
