package main

import (
	"context"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/mw"
	"github.com/gojekfarm/ziggurat/mw/statsd"
)

const RoutePlainTextLog = "plain-text-log"

func main() {
	app := ziggurat.NewApp()
	router := ziggurat.NewRouter()
	statsdClient := statsd.NewClient(statsd.WithPrefix("super-app"))

	router.HandleFunc(RoutePlainTextLog, func(event ziggurat.MessageEvent, ctx context.Context) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})

	rmw := router.Compose(mw.ProcessingStatusLogger, statsdClient.PublishKafkaLag, statsdClient.PublishHandlerMetrics)

	app.OnStart(func(ctx context.Context, routeNames []string) {
		statsdClient.Run(ctx)
	})

	app.OnStop(func() {

	})

	<-app.Run(context.Background(), rmw,
		ziggurat.Routes{
			RoutePlainTextLog: {
				InstanceCount:    3,
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				GroupID:          "plain_text_consumer",
			},
		})
}
