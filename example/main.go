package main

import (
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/mw"
	"github.com/gojekfarm/ziggurat/mw/rabbitmq"
	"github.com/gojekfarm/ziggurat/mw/statsd"
	"github.com/gojekfarm/ziggurat/server"
)

const RoutePlainTextLog = "plain-text-log"
const RouteJSONLog = "json-log"

type JSONLog struct {
	Value int `json:"value"`
}

func main() {
	app := ziggurat.NewApp()
	router := ziggurat.NewRouter()
	statsdClient := statsd.NewStatsD(statsd.WithPrefix("super-app"))
	httpServer := server.NewHTTPServer()
	rmq := rabbitmq.NewRabbitRetrier(app.Context(),
		[]string{"amqp://user:bitnami@localhost:5672/"},
		rabbitmq.QueueConfig{RouteJSONLog: {DelayQueueExpirationInMS: "500", RetryCount: 2}}, nil)

	router.HandleFunc(RoutePlainTextLog, func(event ziggurat.Event, app ziggurat.AppContext) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})

	router.HandleFunc(RouteJSONLog, func(event ziggurat.Event, app ziggurat.AppContext) ziggurat.ProcessStatus {
		jsonLog := &JSONLog{}
		if err := ziggurat.JSON(event.MessageValue(), jsonLog); err != nil {
			ziggurat.LogError(err, "json decode error", nil)
			return ziggurat.SkipMessage
		}
		if jsonLog.Value%2 == 0 {
			return ziggurat.RetryMessage
		}
		return ziggurat.ProcessingSuccess
	})

	rmw := router.Compose(mw.ProcessingStatusLogger, statsdClient.PublishKafkaLag, statsdClient.PublishHandlerMetrics, rmq.Retrier)

	app.OnStart(func(a ziggurat.AppContext) {
		statsdClient.Start(app)
		httpServer.Start(app)
		rmq.StartConsumers(app, app.Handler())
	})

	app.OnStop(func(a ziggurat.AppContext) {
		httpServer.Stop(a)
		statsdClient.Stop()
	})

	<-app.Run(rmw, ziggurat.Routes{
		RoutePlainTextLog: ziggurat.Stream{
			InstanceCount:    1,
			BootstrapServers: "localhost:9092",
			OriginTopics:     "plain-text-log",
			GroupID:          "plain_text_consumer",
		},
		RouteJSONLog: ziggurat.Stream{
			InstanceCount:    1,
			BootstrapServers: "localhost:9092",
			OriginTopics:     "json-log",
			GroupID:          "json_consumer",
		},
	})
}
