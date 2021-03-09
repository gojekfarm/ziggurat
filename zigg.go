package ziggurat

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/gojekfarm/ziggurat/logger"
)

type StartFunction func(ctx context.Context)
type StopFunction func()

type Ziggurat struct {
	handler   Handler
	Logger    StructuredLogger
	startFunc StartFunction
	stopFunc  StopFunction
	streams   Streamer
}

func (z *Ziggurat) Run(ctx context.Context, streams Streamer, handler Handler) error {

	if z.Logger == nil {
		z.Logger = logger.NewJSONLogger("info")
	}

	if streams == nil {
		panic("`streams` cannot be nil")
	}

	if handler == nil {
		panic("`handler` cannot be nil")
	}

	z.streams = streams

	parentCtx, canceler := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	z.handler = handler

	err := <-z.start(parentCtx, z.startFunc)
	z.Logger.Error("streams shutdown", err)
	canceler()

	z.stop(z.stopFunc)

	return err
}

func (z *Ziggurat) start(ctx context.Context, startCallback StartFunction) chan error {
	if startCallback != nil {
		z.Logger.Info("invoking start function")
		startCallback(ctx)
	}

	streamsStop := z.streams.Stream(ctx, z.handler)
	return streamsStop
}

func (z *Ziggurat) stop(stopFunc StopFunction) {
	if stopFunc != nil {
		z.Logger.Info("invoking stop function")
		stopFunc()
	}
}

func (z *Ziggurat) StartFunc(function StartFunction) {
	z.startFunc = function
}

func (z *Ziggurat) StopFunc(function StopFunction) {
	z.stopFunc = function
}
