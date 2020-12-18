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
	msgLogger := mw.NewProcessingStatusLogger(nil)

	router.HandleFunc(RoutePlainTextLog, func(event ziggurat.MessageEvent, ctx context.Context) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})

	rmw := router.Compose(msgLogger.Log)

	app.OnStart(func(ctx context.Context, routeNames []string) {
	})

	<-app.Run(context.Background(), rmw, ziggurat.Routes{
		RoutePlainTextLog: {
			InstanceCount:    2,
			BootstrapServers: "localhost:9092",
			OriginTopics:     "plain-text-log",
			GroupID:          "plain_text_consumer",
		},
	})
}
