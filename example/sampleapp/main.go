package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/gojekfarm/ziggurat/mw/statsd"

	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
	"github.com/gojekfarm/ziggurat/mw/rabbitmq"
)

func main() {
	var zig ziggurat.Ziggurat
	router := ziggurat.NewRouter()

	ctx := context.Background()
	l := logger.NewLogger(logger.LevelInfo)

	statsClient := statsd.NewPublisher(statsd.WithLogger(l),
		statsd.WithDefaultTags(statsd.StatsDTag{"ziggurat-version": "v162"}),
		statsd.WithPrefix("ziggurat_v162"))

	kcg := kafka.ConsumerGroup{
		Logger: nil,
		GroupConfig: kafka.ConsumerConfig{
			BootstrapServers: "localhost:9092",
			GroupID:          "foo.id",
			Topics:           []string{"foo"},
		},
	}

	ar := rabbitmq.AutoRetry([]rabbitmq.QueueConfig{
		{
			QueueKey:              "plain_text_messages_retry",
			DelayExpirationInMS:   "500",
			ConsumerPrefetchCount: 1,
			ConsumerCount:         1,
			RetryCount:            1,
		},
	}, rabbitmq.WithLogger(l),
		rabbitmq.WithUsername("user"),
		rabbitmq.WithPassword("bitnami"),
		rabbitmq.WithConnectionTimeout(3*time.Second))

	router.HandlerFunc("cpool/", func(ctx context.Context, event *ziggurat.Event) error {
		if rand.Intn(1000)%2 == 0 {
			l.Info("retrying")
			err := ar.Retry(ctx, event, "plain_text_messages_retry")
			l.Info("retrying finished")
			return err
		}
		return nil
	})

	h := ziggurat.Use(router, statsClient.PublishEventDelay, statsClient.PublishHandlerMetrics)

	if runErr := zig.Run(ctx, h, &kcg); runErr != nil {
		l.Error("error running streams", runErr)
	}

}
