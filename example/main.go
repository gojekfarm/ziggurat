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
	l := logger.NewJSONLogger(logger.LevelInfo)

	ar := rabbitmq.AutoRetry(
		[]rabbitmq.QueueConfig{{
			QueueName:           "pt_log",
			DelayExpirationInMS: "2000",
			RetryCount:          2,
			WorkerCount:         5,
		}},
		rabbitmq.WithPassword("bitnami"),
		rabbitmq.WithUsername("user"),
		rabbitmq.WithLogger(l))

	kafkaStreams := kafka.Streams{
		StreamConfig: kafka.StreamConfig{
			{
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				ConsumerGroupID:  "another_brick_in_the_wall",
				ConsumerCount:    2,
			},
		},
		Logger: l,
	}

	r.HandleFunc("localhost:9092/another_brick_in_the_wall/",
		ar.Wrap(func(ctx context.Context, event *ziggurat.Event) error {
			return ziggurat.Retry
		}, "pt_log"))

	zig.StartFunc(func(ctx context.Context) {
		l.Error("error starting rabbitmq publishers", ar.InitPublishers(ctx))
	})

	if runErr := zig.RunAll(ctx, &r, &kafkaStreams, ar); runErr != nil {
		l.Error("", runErr)
	}
}
