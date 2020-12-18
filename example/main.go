package main

import (
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/mw"
	"github.com/gojekfarm/ziggurat/mw/statsd"
)

const RoutePlainTextLog = "plain-text-log"

func main() {
	app := ziggurat.NewApp()
	router := ziggurat.NewRouter()
	statsdClient := statsd.NewClient(statsd.WithPrefix("super-app"))

	router.HandleFunc(RoutePlainTextLog, func(event ziggurat.MessageEvent, z *ziggurat.Ziggurat) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})

	rmw := router.Compose(mw.ProcessingStatusLogger, statsdClient.PublishKafkaLag, statsdClient.PublishHandlerMetrics)

	app.OnStart(func(z *ziggurat.Ziggurat) {
		statsdClient.Start(z)
	})

	app.OnStop(func(z *ziggurat.Ziggurat) {
		statsdClient.Stop()
	})

	<-app.Run(rmw, ziggurat.Routes{
		RoutePlainTextLog: {
			InstanceCount:    3,
			BootstrapServers: "localhost:9092",
			OriginTopics:     "plain-text-log",
			GroupID:          "plain_text_consumer",
		},
	})
}
