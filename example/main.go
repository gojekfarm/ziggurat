package main

import (
	"context"
	"errors"
	"github.com/gojekfarm/ziggurat/mw/event"

	"github.com/gojekfarm/ziggurat/mw/prometheus"
	"github.com/gojekfarm/ziggurat/mw/statsd"

	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
	"github.com/gojekfarm/ziggurat/router"
)

func main() {
	var zig ziggurat.Ziggurat
	jsonLogger := logger.NewJSONLogger(logger.LevelInfo)
	statsdPublisher := statsd.NewPublisher(statsd.WithLogger(jsonLogger))
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

	r.HandleFunc("plain-text-log", func(ctx context.Context, event *ziggurat.Event) interface{} {
		return nil
	})

	r.HandleFunc("json-log", func(ctx context.Context, event *ziggurat.Event) interface{} {
		return errors.New("could not process message")
	})

	handler := r.Compose(
		event.Logger(jsonLogger),
		statsdPublisher.PublishKafkaLag,
		statsdPublisher.PublishHandlerMetrics,
		prometheus.PublishHandlerMetrics,
	)

	promStop := make(chan struct{}, 1)

	zig.StartFunc(func(ctx context.Context) {
		go func() {
			jsonLogger.Error("", prometheus.StartMonitoringServer(ctx))
			promStop <- struct{}{}
		}()

		jsonLogger.Error("could not start statsd publisher", statsdPublisher.Run(ctx))
		prometheus.Register()
	})

	if runErr := zig.Run(ctx, kafkaStreams, handler); runErr != nil {
		jsonLogger.Error("could not start streams", runErr)
	}

	<-promStop
}
