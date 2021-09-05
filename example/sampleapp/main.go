//go:build ignore
// +build ignore

package main

import (
	"context"
	"net/http"

	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
	"github.com/gojekfarm/ziggurat/mw/rabbitmq"
	"github.com/gojekfarm/ziggurat/mw/statsd"
	"github.com/gojekfarm/ziggurat/server"
	"github.com/julienschmidt/httprouter"
)

func main() {
	var zig ziggurat.Ziggurat
	var r kafka.Router

	ctx := context.Background()

	l := logger.NewLogger(logger.LevelInfo)
	s := statsd.NewPublisher(statsd.WithPrefix("example_go_ziggurat"),
		statsd.WithDefaultTags(statsd.StatsDTag{"app_name": "example_go_ziggurat"}),
		statsd.WithLogger(l))

	srvr := server.NewHTTPServer()

	ar := rabbitmq.AutoRetry([]rabbitmq.QueueConfig{
		{
			QueueName:             "pt_retries",
			DelayExpirationInMS:   "1000",
			RetryCount:            5,
			ConsumerPrefetchCount: 10,
			ConsumerCount:         10,
		}},
		rabbitmq.WithLogger(l),
		rabbitmq.WithUsername("user"),
		rabbitmq.WithPassword("bitnami"))

	srvr.ConfigureHTTPEndpoints(func(r *httprouter.Router) {
		r.Handler(http.MethodGet, "/dead_set", ar.DSViewHandler(ctx))
		r.Handler(http.MethodGet, "/dead_set/replay", ar.DSReplayHandler(ctx))
	})

	kafkaStreams := kafka.Streams{
		StreamConfig: kafka.StreamConfig{
			{
				BootstrapServers: "g-gojek-id-mainstream.golabs.io:6668",
				OriginTopics:     "^.*-booking-log",
				ConsumerGroupID:  "booking_log_consumer_go",
				ConsumerCount:    2,
			},
		},
		Logger: l,
	}

	r.HandleFunc("g-gojek-id-mainstream.golabs.io:6668/booking_log_consumer_go/", ar.Wrap(func(ctx context.Context, event *ziggurat.Event) error {
		return ziggurat.Retry
	}, "pt_retries"))

	h := ziggurat.Use(&r, s.PublishHandlerMetrics, s.PublishEventDelay)

	done := make(chan struct{})
	zig.StartFunc(func(ctx context.Context) {
		err := s.Run(ctx)
		l.Error("error running statsd publisher", err)

		err = ar.InitPublishers(ctx)
		if err != nil {
			panic("could not start publishers:" + err.Error())
		}

		go func() {
			err := srvr.Run(ctx)
			l.Error("could not start http server", err)
			done <- struct{}{}

		}()

	})

	if runErr := zig.RunAll(ctx, h, &kafkaStreams, ar); runErr != nil {
		l.Error("error running streams", runErr)
	}

	<-done

}
