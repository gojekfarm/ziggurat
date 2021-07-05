//+build ignore

package main

import (
	"context"
	"fmt"

	"github.com/gojekfarm/ziggurat/mw/statsd"

	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
)

func main() {
	var zig ziggurat.Ziggurat
	var r kafka.Router

	jsonLogger := logger.NewJSONLogger(logger.LevelInfo)
	ctx := context.Background()
	statsdPub := statsd.NewPublisher(
		statsd.WithLogger(jsonLogger),
		statsd.WithDefaultTags(statsd.StatsDTag{"app_name": "example_app"}),
	)

	kafkaStreams := kafka.Streams{
		StreamConfig: kafka.StreamConfig{
			{
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				ConsumerGroupID:  "plain_text_consumer",
				ConsumerCount:    1,
			},
		},
		Logger: jsonLogger,
	}

	r.HandleFunc("localhost:9092/plain_text_consumer/.*-text-log/0$", func(ctx context.Context, event *ziggurat.Event) error {
		fmt.Println("received message ", string(event.Value), " on partition 0")
		return nil
	})

	r.HandleFunc("localhost:9092/plain_text_consumer/.*-text-log/1$", func(ctx context.Context, event *ziggurat.Event) error {
		fmt.Println("received message ", string(event.Value), " on partition 1")
		return nil
	})

	zig.StartFunc(func(ctx context.Context) {
		jsonLogger.Error("error running statsd publisher", statsdPub.Run(ctx))
	})

	if runErr := zig.Run(ctx, &kafkaStreams, &r); runErr != nil {
		jsonLogger.Error("could not start streams", runErr)
	}
}
