package dsl

import (
	"context"

	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/logger"
	"github.com/gojekfarm/ziggurat/router"
)

type Routes map[string]ziggurat.HandlerFunc
type Middleware []func(handler ziggurat.Handler) ziggurat.Handler

type App struct {
	Streams    ziggurat.Streamer
	Routes     Routes
	Middleware Middleware
	Logger     ziggurat.StructuredLogger
	StartFunc  func(ctx context.Context)
}

func (a *App) Run(ctx context.Context) error {
	if a.Logger == nil {
		a.Logger = logger.NewJSONLogger(logger.LevelInfo)
	}
	if a.Streams == nil {
		panic("Streams cannot be nil")
	}

	if a.Routes == nil || len(a.Routes) < 1 {
		panic("Routes cannot be empty")
	}

	if a.StartFunc == nil {
		a.StartFunc = func(ctx context.Context) {}
	}
	var z ziggurat.Ziggurat
	r := router.New()

	for routeName, handler := range a.Routes {
		r.HandleFunc(routeName, handler)
	}

	handler := r.Compose(a.Middleware...)

	return z.Run(ctx, a.Streams, handler)

}
