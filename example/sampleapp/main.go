package main

import (
	"context"
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
		StreamConfig: kafka.StreamConfig{
			{
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				ConsumerGroupID:  "text_consumer",
				ConsumerCount:    1,
			},
		},
		Logger: l,
	}

	r.HandleFunc("localhost:9092/text_consumer/", func(ctx context.Context, event *ziggurat.Event) error {
		return nil
	})

	if runErr := zig.RunAll(ctx, &r, &ks); runErr != nil {
		l.Error("error running streams", runErr)
	}

}
