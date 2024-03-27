package main

import (
	"context"
	"fmt"
	"github.com/gojekfarm/ziggurat/v2"
	"github.com/gojekfarm/ziggurat/v2/kafka"
	"github.com/gojekfarm/ziggurat/v2/logger"
)

func main() {
	var zig ziggurat.Ziggurat

	ctx := context.Background()
	l := logger.NewLogger(logger.LevelInfo)

	kcg := kafka.ConsumerGroup{
		Logger: logger.NewLogger(logger.LevelInfo),
		GroupConfig: kafka.ConsumerConfig{
			BootstrapServers: "g-gojek-id-mainstream.golabs.io:6668",
			GroupID:          "foo.id",
			ConsumerCount:    1,
			Topics:           []string{"^.*-booking-log"},
		},
	}

	if runErr := zig.Run(ctx, ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) {
		fmt.Println("path:", event.RoutingPath)
	}), &kcg); runErr != nil {
		l.Error("error running consumers", runErr)
	}

}
