package statsd

import (
	"context"
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

// NewPublisher creates a new publisher with an embedded statsd client
// use statsd.WithPrefix to specify a prefix which will be sent as a common label with all metrics
// defaults to "ziggurat_statsd"
// use statsd.WithHost to specify a custom host:port string, defaults to localhost:8125
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

// Run methods runs the publisher and starts up the go-routine publisher in the background
// the go-routine publisher publishes the go-routine count every 10 seconds
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

// IncCounter increments a counter "metric_name|c"
// returns a publish err on failure
func (s *Client) IncCounter(metricName string, value int64, arguments map[string]string) error {
	tags := constructTags(arguments)
	finalMetricName := s.constructFullMetricStr(metricName, tags)

	return s.client.Inc(finalMetricName, value, 1.0)
}

// Gauge publishes a metric of type gauge "metric_name|g"
// returns a publish error on failure
func (s *Client) Gauge(metricName string, value int64, arguments map[string]string) error {
	tags := constructTags(arguments)
	finalMetricName := s.constructFullMetricStr(metricName, tags)
	return s.client.Gauge(finalMetricName, value, 1.0)
}

// PublishHandlerMetrics is a ziggurat middleware which publishes the
// handler_execution_time - time taken for the handler func to execute in milliseconds
// processing_failure_count - count of errors returned by the handler func
// processing_success_count - count of nil errors returned by handler
// message_count - count of all messages encountered by the handler
func (s *Client) PublishHandlerMetrics(handler ziggurat.Handler) ziggurat.Handler {
	f := func(ctx context.Context, event *ziggurat.Event) error {
		t1 := time.Now()
		err := handler.Handle(ctx, event)
		args := map[string]string{
			"route": event.Path,
		}
		s.Logger.Error(publishErrMsg, s.Gauge("handler_execution_time", time.Since(t1).Milliseconds(), args))
		s.Logger.Error(publishErrMsg, s.IncCounter("message_count", 1, args))

		if err != nil && err == ziggurat.Retry {
			s.Logger.Error(publishErrMsg, s.IncCounter("event_retry_count", 1, args))
		}

		if err != nil {
			s.Logger.Error(publishErrMsg, s.IncCounter("processing_failure_count", 1, args))
		}
		s.Logger.Error(publishErrMsg, s.IncCounter("processing_success_count", 1, args))
		return err
	}
	return ziggurat.HandlerFunc(f)
}

// PublishKafkaLag publishes the kafka lag per topic in milliseconds
// kafka_lag - time difference in milliseconds between the kafka event timestamp and the current time
func (s *Client) PublishKafkaLag(handler ziggurat.Handler) ziggurat.Handler {
	f := func(ctx context.Context, event *ziggurat.Event) error {
		headers := event.Headers
		args := map[string]string{}

		args["topic"] = headers["x-kafka-topic"]
		args["partition"] = headers["x-kafka-partition"]

		diff := event.ReceivedTimestamp.Sub(event.ProducerTimestamp).Milliseconds()
		s.Logger.Error(publishErrMsg, s.Gauge("kafka_lag", diff, args))
		return handler.Handle(ctx, event)

	}
	return ziggurat.HandlerFunc(f)
}
