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
		Logger: l,
		GroupConfig: kafka.ConsumerConfig{
			BootstrapServers: "g-gojek-id-mainstream.golabs.io:6668",
			GroupID:          "bar.id",
			ConsumerCount:    1,
			AutoOffsetReset:  "earliest",
			Topics:           []string{"^.*-booking-log"},
		},
	}

	router := ziggurat.NewRouter()
	router.HandlerFunc("bar.id/.*", func(ctx context.Context, event *ziggurat.Event) {
		fmt.Println("path:", event.RoutingPath)
	})

	if runErr := zig.Run(ctx, router, &kcg); runErr != nil {
		l.Error("error running streams", runErr)
	}

}
