//+build ignore

package main

import (
	"context"

	"github.com/gojekfarm/ziggurat/mw/event"

	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
	"github.com/gojekfarm/ziggurat/router"
)

func main() {
	var zig ziggurat.Ziggurat
	jsonLogger := logger.NewJSONLogger(logger.LevelInfo)
	ctx := context.Background()

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
		Logger: jsonLogger,
	}

	r := router.New()

	r.HandleFunc("plain-text-log", func(ctx context.Context, event *ziggurat.Event) error {
		return nil
	})

	handler := r.Compose(event.Logger(jsonLogger))

	if runErr := zig.Run(ctx, &kafkaStreams, handler); runErr != nil {
		jsonLogger.Error("could not start streams", runErr)
	}

}
