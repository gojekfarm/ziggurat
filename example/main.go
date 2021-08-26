//+build ignore

package main

import (
	"context"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
	"github.com/gojekfarm/ziggurat/mw/rabbitmq"
	"github.com/gojekfarm/ziggurat/mw/statsd"
	"github.com/gojekfarm/ziggurat/server"
	"github.com/julienschmidt/httprouter"
	"net/http"
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

	ar := rabbitmq.AutoRetry([]rabbitmq.QueueConfig{{
		QueueName:           "pt_retries",
		DelayExpirationInMS: "3000",
		RetryCount:          5,
		WorkerCount:         10,
	}}, rabbitmq.WithLogger(l),
		rabbitmq.WithUsername("user"),
		rabbitmq.WithPassword("bitnami"))

	srvr.ConfigureHTTPEndpoints(func(r *httprouter.Router) {
		r.Handler(http.MethodGet, "/dead_set", ar.DSViewHandler(ctx))
		r.Handler(http.MethodGet, "/dead_set/replay", ar.DSReplayHandler(ctx))
	})

	kafkaStreams := kafka.Streams{
		StreamConfig: kafka.StreamConfig{
			{
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				ConsumerGroupID:  "another_brick_in_the_wall",
				ConsumerCount:    2,
			},
		},
		Logger: l,
	}

	r.HandleFunc("localhost:9092/another_brick_in_the_wall/", ar.Wrap(func(ctx context.Context, event *ziggurat.Event) error {
		return ziggurat.Retry
	}, "pt_retries"))

	h := r.Compose(s.PublishHandlerMetrics)

	done := make(chan struct{})
	zig.StartFunc(func(ctx context.Context) {
		err := s.Run(ctx)
		l.Error("", err)

		err = ar.InitPublishers(ctx)
		l.Error("", err)

		go func() {
			srvr.Run(ctx)
			done <- struct{}{}
		}()

	})

	if runErr := zig.RunAll(ctx, h, &kafkaStreams,ar); runErr != nil {
		l.Error("", runErr)
	}

	<-done

}
