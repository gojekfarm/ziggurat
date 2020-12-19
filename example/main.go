package main

import (
	"context"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/mw"
	"github.com/gojekfarm/ziggurat/mw/metrics"
	"github.com/gojekfarm/ziggurat/mw/retry"
	"github.com/gojekfarm/ziggurat/server"
)

const RoutePlainTextLog = "plain-text-log"

func main() {
	app := ziggurat.NewApp()
	router := ziggurat.NewRouter()
	loggerMW := mw.NewProcessingStatusLogger()
	retryMW := retry.NewRabbitRetrier(
		[]string{"amqp://user:bitnami@localhost:5672"},
		retry.QueueConfig{
			RoutePlainTextLog: {DelayQueueExpirationInMS: "500", RetryCount: 2},
		},
		nil)
	statsDClient := metrics.NewClient(metrics.WithPrefix("super-app"))
	httpServer := server.NewHTTPServer()

	router.HandleFunc(RoutePlainTextLog, func(event *ziggurat.Message, ctx context.Context) ziggurat.ProcessStatus {
		return ziggurat.RetryMessage
	})

	handler := router.Compose(loggerMW.LogStatus, statsDClient.PublishKafkaLag, statsDClient.PublishHandlerMetrics, retryMW.Retrier)

	app.StartFunc(func(ctx context.Context, routeNames []string) {
		retryMW.RunPublisher(ctx)
		retryMW.RunConsumers(ctx, handler)
		statsDClient.Run(ctx)
		httpServer.Run(ctx)
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
