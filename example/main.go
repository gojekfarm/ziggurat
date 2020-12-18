package main

import (
	"context"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/mw"
)

const RoutePlainTextLog = "plain-text-log"

func main() {
	app := ziggurat.NewApp()
	router := ziggurat.NewRouter()
	loggerMW := mw.NewProcessingStatusLogger()

	router.HandleFunc(RoutePlainTextLog, func(event ziggurat.MessageEvent, ctx context.Context) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})

	app.OnStart(func(ctx context.Context, routeNames []string) {
	})

	<-app.Run(context.Background(), loggerMW.LogStatus(router),
		ziggurat.Routes{
			RoutePlainTextLog: {
				InstanceCount:    2,
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				GroupID:          "plain_text_consumer",
			},
		})
}
