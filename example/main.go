package main

import (
	"context"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/mw"
)

const RoutePlainTextLog = "plain-text-log"

func main() {
	app := &ziggurat.Ziggurat{}
	router := ziggurat.NewRouter()
	statusLogger := mw.NewProcessingStatusLogger()

	router.HandleFunc(RoutePlainTextLog, func(event ziggurat.Message, ctx context.Context) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})

	handler := router.Compose(statusLogger.LogStatus)

	app.StartFunc(func(ctx context.Context) {

	})

	<-app.Run(context.Background(), handler,
		ziggurat.Routes{
			RoutePlainTextLog: {
				ziggurat.KafkaKeyBootstrapServers: "localhost:9092",
				ziggurat.KafkaKeyOriginTopics:     "plain-text-log",
				ziggurat.KafkaKeyConsumerGroup:    "plain_text_consumer",
				ziggurat.KafkaKeyConsumerCount:    2,
			},
		})
}
