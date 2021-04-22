package ziggurat

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/gojekfarm/ziggurat/logger"
)

type StartFunction func(ctx context.Context)
type StopFunction func()

type ErrRunAll struct {
	errors []error
}

func (e ErrRunAll) Error() string {
	return "one of the streams errored out, use the Inspect method to know more"
}

func (e ErrRunAll) Inspect() []error {
	return e.errors
}

// Ziggurat serves as a container for streams to run in
// can be used without initialization
// var z ziggurat.Ziggurat
// z.Run(ctx context.Context,s ziggurat.Streamer,h ziggurat.Handler)
type Ziggurat struct {
	handler   Handler
	Logger    StructuredLogger
	startFunc StartFunction
	stopFunc  StopFunction
	streams   Streamer
}

// Run method runs the provided streams and blocks on it until an error is encountered
// run multiple streams concurrently by wrapping the Run method in a go-routine
// the context provided is notified on SIGINT and SIGTERM
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

	err := z.start(parentCtx, z.startFunc)
	z.Logger.Error("streams shutdown", err)
	canceler()

	z.stop(z.stopFunc)

	return err
}

func (z *Ziggurat) start(ctx context.Context, startCallback StartFunction) error {
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

// StartFunc is used to start additional application states
// advantage of using the StartFunc is that the context remains consistent between
// your streams and other additional application states like database(s) REST API client(s)
func (z *Ziggurat) StartFunc(function StartFunction) {
	z.startFunc = function
}

// StopFunc is immediately called after streams are shutdown
// is used to perform cleanup operations
func (z *Ziggurat) StopFunc(function StopFunction) {
	z.stopFunc = function
}
