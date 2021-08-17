package main

import (
	"context"
	"github.com/gojekfarm/ziggurat/mw/rabbitmq"

	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
)

func main() {
	var zig ziggurat.Ziggurat
	var r kafka.Router

	ctx := context.Background()
	l := logger.NewLogger(logger.LevelInfo)

	ar := rabbitmq.AutoRetry(
		rabbitmq.WithPassword("bitnami"),
		rabbitmq.WithUsername("user"),
		rabbitmq.WithLogger(l))

	kafkaStreams := kafka.Streams{
		StreamConfig: kafka.StreamConfig{
			{
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				ConsumerGroupID:  "plain_text_consumer",
				ConsumerCount:    1,
				RouteGroup:       "plain-text-log",
			},
		},
		Logger: l,
	}

	r.HandleFunc("localhost:9092/plain_text_consumer/",
		ar.Wrap(func(ctx context.Context, event *ziggurat.Event) error {
			l.Info("received message", map[string]interface{}{"value": string(event.Value)})
			return ziggurat.Retry
		}, "booking_log"))

	zig.StartFunc(func(ctx context.Context) {
		go func() {
			err := ar.Run(ctx, &r, rabbitmq.WithQueues(
				rabbitmq.QueueConfig{
					QueueName:           "booking_log",
					DelayExpirationInMS: "1000",
					RetryCount:          5,
					WorkerCount:         10,
				}))
			l.Error("error starting rabbitmq", err)
		}()
	})

	if runErr := zig.Run(ctx, &kafkaStreams, &r); runErr != nil {
		l.Error("could not start streams", runErr)
	}
}
