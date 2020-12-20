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
		ziggurat.StreamRoutes{
			RoutePlainTextLog: {
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				ConsumerGroupID:  "plain_text_consumer",
				ConsumerCount:    2,
			},
		})
}
