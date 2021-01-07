package main

import (
	"context"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/logger"
	"github.com/gojekfarm/ziggurat/mw"
	"github.com/gojekfarm/ziggurat/router"
	"github.com/gojekfarm/ziggurat/streams"
)

func main() {
	jsonLogger := logger.NewJSONLogger("disabled")
	kafkaStreams := &streams.Kafka{
		KafkaRouteGroup: streams.KafkaRouteGroup{
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
	statusLogger := mw.NewProcessingStatusLogger(mw.WithLogger(jsonLogger))

	r.HandleFunc("json-log", func(event ziggurat.Event) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})

	r.HandleFunc("plain-text-log", func(event ziggurat.Event) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})

	handler := r.Compose(statusLogger.LogStatus)

	zig := &ziggurat.Ziggurat{Logger: jsonLogger}
	<-zig.Run(context.Background(), kafkaStreams, handler)
}
