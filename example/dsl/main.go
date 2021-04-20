package main

import (
	"context"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/dsl"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
)

func main() {
	l := logger.NewJSONLogger(logger.LevelInfo)
	c := context.Background()
	a := dsl.App{
		Streams: &kafka.Streams{
			Logger: l,
			StreamConfig: kafka.StreamConfig{
				{
					BootstrapServers: "localhost:9092",
					ConsumerCount:    1,
					ConsumerGroupID:  "plain_text_consumer",
					RouteGroup:       "plain-text-log",
					OriginTopics:     "plain-text-log",
				}},
		},
		Routes: dsl.Routes{
			"plain-text-log": func(ctx context.Context, event *ziggurat.Event) error {
				l.Info("received message", map[string]interface{}{"value": string(event.Value)})
				return nil
			},
		},
		Middleware: dsl.Middleware{},
		Logger:     l,
	}
	l.Error("error running app", a.Run(c))
}
