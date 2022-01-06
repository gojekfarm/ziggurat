//go:build ignore
// +build ignore

package main

import (
	"context"
	"strconv"
	"strings"

	"github.com/gojekfarm/ziggurat/mw/prometheus"

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
				Topics:           "plain-text-log",
				GroupID:          "pt_consumer",
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

	wait := make(chan struct{})
	zig.StartFunc(func(ctx context.Context) {
		go func() {
			err := prometheus.StartMonitoringServer(ctx)
			l.Error("error running prom server", err)
			wait <- struct{}{}
		}()
	})

	h := ziggurat.Use(&r, prometheus.PublishHandlerMetrics)

	if runErr := zig.RunAll(ctx, h, &ks, ar); runErr != nil {
		l.Error("error running streams", runErr)
	}

	<-wait
}
