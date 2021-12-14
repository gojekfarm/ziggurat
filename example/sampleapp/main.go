package main

import (
	"context"
	"strconv"
	"strings"

	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
	"github.com/gojekfarm/ziggurat/mw/rabbitmq"
	"github.com/gojekfarm/ziggurat/mw/statsd"
)

func main() {
	var zig ziggurat.Ziggurat
	var r kafka.Router

	statsdPub := statsd.NewPublisher(statsd.WithDefaultTags(map[string]string{
		"app_name": "sample_app",
	}))
	ctx := context.Background()
	l := logger.NewLogger(logger.LevelInfo)

	ar := rabbitmq.AutoRetry(rabbitmq.Queues{{
		QueueName:             "pt_retries",
		DelayExpirationInMS:   "1000",
		RetryCount:            3,
		ConsumerPrefetchCount: 10,
		ConsumerCount:         2,
	}}, rabbitmq.WithUsername("user"),
		rabbitmq.WithLogger(l),
		rabbitmq.WithPassword("bitnami"))

	ks := kafka.Streams{
		StreamConfig: kafka.StreamConfig{
			{
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				ConsumerGroupID:  "pt_consumer",
				ConsumerCount:    2,
				RouteGroup:       "plain-text-messages",
			},
		},
		Logger: l,
	}

	r.HandleFunc("plain-text-messages/", ar.Wrap(func(ctx context.Context, event *ziggurat.Event) error {
		val := string(event.Value)
		s := strings.Split(val, "_")
		num, err := strconv.Atoi(s[1])
		if err != nil {
			return err
		}
		if num%2 == 0 {
			return ziggurat.Retry
		}
		return nil
	}, "pt_retries"))

	zig.StartFunc(func(ctx context.Context) {
		err := statsdPub.Run(ctx)
		l.Error("statsd publisher error", err)
	})

	zig.StopFunc(func() {
		statsdPub.Close()
	})

	if runErr := zig.RunAll(ctx, &r, &ks, ar); runErr != nil {
		l.Error("error running streams", runErr)
	}

}
