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
	var r kafka.Router

	ctx := context.Background()
	l := logger.NewLogger(logger.LevelInfo)

	statsClient := statsd.NewPublisher(statsd.WithLogger(l),
		statsd.WithDefaultTags(statsd.StatsDTag{"ziggurat-version": "v162"}),
		statsd.WithPrefix("ziggurat_v162"))

	ks := kafka.Streams{
		StreamConfig: kafka.StreamConfig{{
			BootstrapServers: "localhost:9092",
			Topics:           "json-log",
			GroupID:          "ziggurat_consumer_local",
			ConsumerCount:    2,
			RouteGroup:       "cpool"}},
		Logger: l,
	}

	ar := rabbitmq.AutoRetry([]rabbitmq.QueueConfig{
		{
			QueueKey:              "plain_text_messages_retry",
			DelayExpirationInMS:   "500",
			ConsumerPrefetchCount: 1,
			ConsumerCount:         1,
			RetryCount:            1,
		},
		{
			QueueKey:      "bulk_cons",
			ConsumerCount: 10,
			Type:          rabbitmq.WorkerQueue,
		},
	}, rabbitmq.WithLogger(l),
		rabbitmq.WithUsername("user"),
		rabbitmq.WithPassword("bitnami"),
		rabbitmq.WithConnectionTimeout(3*time.Second))

	r.HandleFunc("cpool/", func(ctx context.Context, event *ziggurat.Event) error {
		if rand.Intn(1000)%2 == 0 {
			l.Info("retrying")
			err := ar.Retry(ctx, event, "plain_text_messages_retry")
			l.Info("retrying finished")
			return err
		}
		return ar.SendToWorker(ctx, event, "bulk_cons")

	})

	zig.StartFunc(func(ctx context.Context) {
		statsClient.Run(ctx)
	})

	h := ziggurat.Use(&r, statsClient.PublishEventDelay, statsClient.PublishHandlerMetrics)

	if runErr := zig.RunAll(ctx, h, &ks, ar); runErr != nil {
		l.Error("error running streams", runErr)
	}

}
