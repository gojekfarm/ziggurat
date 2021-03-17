//+build ignore

package main

import (
	"context"

	"github.com/gojekfarm/ziggurat/mw"
	"github.com/gojekfarm/ziggurat/mw/statsd"

	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
	"github.com/gojekfarm/ziggurat/router"
)

func main() {

	jsonLogger := logger.NewJSONLogger(logger.LevelInfo)
	statsdPublisher := statsd.NewPublisher()
	psLogger := mw.ProcessingStatusLogger{Logger: jsonLogger}
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
		return ziggurat.ErrProcessingFailed{}
	})

	handler := r.Compose(psLogger.LogStatus, statsdPublisher.PublishKafkaLag, statsdPublisher.PublishHandlerMetrics)

	zig := &ziggurat.Ziggurat{}

	zig.StartFunc(func(ctx context.Context) {
		statsdPublisher.Run(ctx)
	})
	zig.Run(ctx, kafkaStreams, handler)

}
