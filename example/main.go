//+build ignore

package main

import (
	"context"
	"net/http"

	"github.com/gojekfarm/ziggurat/mw/rabbitmq"
	"github.com/gojekfarm/ziggurat/server"
	"github.com/julienschmidt/httprouter"

	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
)

func main() {
	var zig ziggurat.Ziggurat
	var r kafka.Router

	ctx := context.Background()
	l := logger.NewLogger(logger.LevelInfo)
	srvr := server.NewHTTPServer(server.WithAddr(":8080"))

	ar := rabbitmq.AutoRetry(
		[]rabbitmq.QueueConfig{{
			QueueName:           "pt_log",
			DelayExpirationInMS: "2000",
			RetryCount:          2,
			WorkerCount:         5,
		}},
		rabbitmq.WithPassword("bitnami"),
		rabbitmq.WithUsername("user"),
		rabbitmq.WithLogger(l))

	srvr.ConfigureHTTPEndpoints(func(r *httprouter.Router) {
		r.Handler(http.MethodGet, "/dead_set", ar.DSViewHandler(ctx))
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

	r.HandleFunc("localhost:9092/another_brick_in_the_wall/", func(ctx context.Context, event *ziggurat.Event) error {
		return nil
	})

	done := make(chan struct{})
	zig.StartFunc(func(ctx context.Context) {
		go func() {
			err := srvr.Run(ctx)
			l.Error("server error", err)
			done <- struct{}{}
		}()
	})

	if runErr := zig.RunAll(ctx, &r, &kafkaStreams); runErr != nil {
		l.Error("", runErr)
	}
	<-done
}
