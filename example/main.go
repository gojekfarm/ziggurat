package main

import (
	"context"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/mw"
)

func main() {
	app := &ziggurat.Ziggurat{}
	router := ziggurat.NewRouter()
	statusLogger := mw.NewProcessingStatusLogger()

	router.HandleFunc("json-log", func(event ziggurat.Event) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})

	router.HandleFunc("plain-text-log", func(event ziggurat.Event) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})

	handler := router.Compose(statusLogger.LogStatus)

	<-app.Run(context.Background(), handler,
		ziggurat.StreamRoutes{
			"json-log": {
				BootstrapServers: "localhost:9092",
				OriginTopics:     "json-log",
				ConsumerGroupID:  "json_consumer",
				ConsumerCount:    1,
			},
			"plain-text-log": {
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				ConsumerGroupID:  "plain_text_consumer",
				ConsumerCount:    1,
			},
		})
}
