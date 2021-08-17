//+build ignore

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
		rabbitmq.WithQueues(
			rabbitmq.QueueConfig{
				QueueName:           "booking_log",
				DelayExpirationInMS: "1000",
				RetryCount:          5,
				WorkerCount:         50,
			}))

	kafkaStreams := kafka.Streams{
		StreamConfig: kafka.StreamConfig{
			{
				BootstrapServers: "g-gojek-id-mainstream.golabs.io:6668",
				OriginTopics:     "^.*-booking-log",
				ConsumerGroupID:  "another_brick_in_the_wall",
				ConsumerCount:    12,
			},
		},
		Logger: l,
	}

	r.HandleFunc("g-gojek-id-mainstream.golabs.io:6668/another_brick_in_the_wall/",
		ar.Wrap(func(ctx context.Context, event *ziggurat.Event) error {
			return ziggurat.Retry
		}, "booking_log"))

	zig.StartFunc(func(ctx context.Context) {

	})

	if runErr := zig.RunAll(ctx, &r, &kafkaStreams, ar); runErr != nil {
		l.Error("could not start streams", runErr)
	}
}
