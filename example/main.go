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
				ConsumerGroupID:  "json_log_consumer",
				ConsumerCount:    2,
				RouteGroup:       "json-log",
			},
		},
		Logger: jsonLogger,
	}
	r := router.New()

	r.HandleFunc("plain-text-log", func(event ziggurat.Event) error {
		return nil
	})

	r.HandleFunc("json-log", func(event ziggurat.Event) error {
		return nil
	})

	processingLogger := &mw.ProcessingStatusLogger{Logger: jsonLogger}

	handler := r.Compose(processingLogger.LogStatus)

	zig := &ziggurat.Ziggurat{Logger: jsonLogger}
	<-zig.Run(context.Background(), kafkaStreams, handler)
}
