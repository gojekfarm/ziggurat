//+build ignore

package main

import (
	"context"

	"github.com/gojekfarm/ziggurat/mw/statsd"

	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
)

func main() {
	var zig ziggurat.Ziggurat
	var r kafka.Router

	l := logger.NewLogger(logger.LevelInfo)
	ctx := context.Background()
	statsdPub := statsd.NewPublisher(
		statsd.WithLogger(l),
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
		Logger: l,
	}

	r.HandleFunc("localhost:9092/plain_text_consumer/", func(ctx context.Context, event *ziggurat.Event) error {
		l.Info("received message", map[string]interface{}{"value": string(event.Value)})
		return nil
	})

	zig.StartFunc(func(ctx context.Context) {
		l.Error("error running statsd publisher", statsdPub.Run(ctx))
	})

	if runErr := zig.Run(ctx, &kafkaStreams, &r); runErr != nil {
		l.Error("could not start streams", runErr)
	}
}
