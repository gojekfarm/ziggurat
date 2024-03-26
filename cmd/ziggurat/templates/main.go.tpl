package main

import (
	"context"


	"github.com/gojekfarm/ziggurat/v2"
	"github.com/gojekfarm/ziggurat/v2/kafka"
	"github.com/gojekfarm/ziggurat/v2/logger"
)

func main() {
	var zig ziggurat.Ziggurat
	router := ziggurat.NewRouter()

	ctx := context.Background()
	l := logger.NewLogger(logger.LevelInfo)


	kcg := kafka.ConsumerGroup{
		Logger: l,
		GroupConfig: kafka.ConsumerConfig{
			BootstrapServers: "localhost:9092",
			GroupID:          "foo.id",
			Topics:           []string{"foo"},
		},
	}

	router.HandlerFunc("cpool/", func(ctx context.Context, event *ziggurat.Event) {
	})


	if runErr := zig.Run(ctx, router, &kcg); runErr != nil {
		l.Error("error running streams", runErr)
	}

}
