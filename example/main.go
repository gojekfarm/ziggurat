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
		RouteGroup: kafka.RouteGroup{
			"json-log": {
				BootstrapServers: "localhost:9092",
				OriginTopics:     "json-log",
				ConsumerGroupID:  "json_consumer",
				ConsumerCount:    2,
			},
			"plain-text-log": {
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				ConsumerGroupID:  "plain_text_consumer",
				ConsumerCount:    1,
			},
		},
		Logger: jsonLogger,
	}
	r := router.New()

	r.HandleFunc("json-log", func(event ziggurat.Event) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})

	r.HandleFunc("plain-text-log", func(event ziggurat.Event) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})

	processingLogger := &mw.ProcessingStatusLogger{Logger: jsonLogger}

	handler := r.Compose(processingLogger.LogStatus)

	zig := &ziggurat.Ziggurat{Logger: jsonLogger}
	<-zig.Run(context.Background(), kafkaStreams, handler)
}
