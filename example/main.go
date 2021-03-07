//+build ignore

package main

import (
	"context"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
	"github.com/gojekfarm/ziggurat/mw"
	"github.com/gojekfarm/ziggurat/router"
)

func main() {

	jsonLogger := logger.NewJSONLogger("info")
	ctx := context.Background()

	kafkaStreams := &kafka.Streams{
		StreamConfig: kafka.StreamConfig{
			{
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				ConsumerGroupID:  "plain_text_consumer",
				ConsumerCount:    1,
				RouteGroup:       "plain-text-log",
			},
			{
				BootstrapServers: "localhost:9092",
				OriginTopics:     "json-log",
				ConsumerGroupID:  "json_consumer",
				ConsumerCount:    1,
				RouteGroup:       "json-log",
			},
		},
		Logger: jsonLogger,
	}
	r := router.New()

	r.HandleFunc("plain-text-log", func(ctx context.Context, event ziggurat.Event) error {
		return nil
	})

	r.HandleFunc("json-log", func(ctx context.Context, event ziggurat.Event) error {
		return ziggurat.ErrProcessingFailed{"retry"}
	})

	handler := mw.NewProcessStatusLogger(r)
	zig := &ziggurat.Ziggurat{Logger: jsonLogger}

	<-zig.Run(ctx, kafkaStreams, handler)

}
