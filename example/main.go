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
	loggerMW := mw.NewProcessingStatusLogger()

	router.HandleFunc(RoutePlainTextLog, func(event *ziggurat.Message, ctx context.Context) ziggurat.ProcessStatus {
		return ziggurat.RetryMessage
	})

	handler := router.Compose(loggerMW.LogStatus)

	app.StartFunc(func(ctx context.Context) {

	})

	<-app.Run(context.Background(), handler,
		ziggurat.Routes{
			RoutePlainTextLog: {
				InstanceCount:    2,
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				GroupID:          "plain_text_consumer",
			},
		})
}
