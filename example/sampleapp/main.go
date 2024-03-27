package main

import (
	"context"
	"fmt"
	"github.com/gojekfarm/ziggurat/v2"
	"github.com/gojekfarm/ziggurat/v2/kafka"
	"github.com/gojekfarm/ziggurat/v2/logger"
	"github.com/gojekfarm/ziggurat/v2/mw/rabbitmq"
)

func main() {
	var zig ziggurat.Ziggurat
	router := ziggurat.NewRouter()

	ctx := context.Background()
	l := logger.NewLogger(logger.LevelInfo)

	kcg := kafka.ConsumerGroup{
		Logger: logger.NewLogger(logger.LevelInfo),
		GroupConfig: kafka.ConsumerConfig{
			BootstrapServers: "g-gojek-id-mainstream-golabs.io:6668",
			GroupID:          "foo.id",
			Topics:           []string{".*-booking-log"},
		},
	}

	router.HandlerFunc("foo.id/*", func(ctx context.Context, event *ziggurat.Event) {
		if rabbitmq.RetryCountFor(event) > 0 {
			fmt.Println("message has been retried")
		} else {
			fmt.Println("new message")
		}
	})

	h := ziggurat.Use(router)

	if runErr := zig.Run(ctx, h, &kcg); runErr != nil {
		l.Error("error running consumers", runErr)
	}

}
