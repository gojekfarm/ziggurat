//go:build ignore
// +build ignore

package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gojekfarm/ziggurat/mw/rabbitmq"
	"github.com/gojekfarm/ziggurat/server"

	"github.com/gojekfarm/ziggurat/mw/prometheus"

	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
)

func main() {
	var zig ziggurat.Ziggurat
	var r kafka.Router

	ctx := context.Background()
	l := logger.NewLogger(logger.LevelInfo)

	ks := kafka.Streams{
		StreamConfig: kafka.StreamConfig{{
			BootstrapServers: "localhost:9092",
			Topics:           "plain-text-log",
			GroupID:          "pt_consumer",
			ConsumerCount:    2,
			RouteGroup:       "plain-text-messages",
		}},
		Logger: l,
	}

	ar := rabbitmq.AutoRetry([]rabbitmq.QueueConfig{
		{
			QueueKey:              "plain_text_messages_retry",
			DelayExpirationInMS:   "5000",
			RetryCount:            5,
			ConsumerPrefetchCount: 1,
			ConsumerCount:         1,
		},
	}, rabbitmq.WithLogger(l),
		rabbitmq.WithUsername("user"),
		rabbitmq.WithPassword("bitnami"),
		rabbitmq.WithConnectionTimeout(3*time.Second))

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
	}, "plain_text_messages_retry"))

	mux := http.NewServeMux()
	mux.Handle("/dead_set", ar.DSReplayHandler(ctx))
	s := http.Server{Addr: ":8080", Handler: mux}

	wait := make(chan struct{})
	zig.StartFunc(func(ctx context.Context) {
		go func() {
			server.Run(ctx, &s)
			wait <- struct{}{}
		}()
	})

	h := ziggurat.Use(&r, prometheus.PublishHandlerMetrics)

	if runErr := zig.RunAll(ctx, h, &ks); runErr != nil {
		l.Error("error running streams", runErr)
	}

	<-wait
}
