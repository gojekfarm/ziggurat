package ziggurat

import (
	"context"
	"fmt"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/gojekfarm/ziggurat/logger"
)

type StartFunction func(ctx context.Context)
type StopFunction func()

// Ziggurat serves as a container for streams to run in
// can be used without initialization
// var z ziggurat.Ziggurat
// z.Run(ctx context.Context,s ziggurat.Streamer,h ziggurat.Handler)
type Ziggurat struct {
	handler      Handler
	Logger       StructuredLogger
	startFunc    StartFunction
	stopFunc     StopFunction
	streams      Streamer
	multiStreams []Streamer
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
	z.handler = handler

	parentCtx, canceler := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

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

// StartFunc is used to start additional application states the
// context remains consistent between
// your streams and other additional application states like database(s) and REST API client(s)
func (z *Ziggurat) StartFunc(function StartFunction) {
	z.startFunc = function
}

// StopFunc is immediately called after streams are shutdown
// is used to perform cleanup operations
func (z *Ziggurat) StopFunc(function StopFunction) {
	z.stopFunc = function
}

// RunAll takes in a handler and multiple streams.
// The streams are started concurrently and the handler is executed for
// all the streams.
func (z *Ziggurat) RunAll(ctx context.Context, handler Handler, streams ...Streamer) error {
	if z.Logger == nil {
		z.Logger = logger.NewJSONLogger(logger.LevelInfo)
	}
	if len(streams) < 1 {
		panic("error: at least one streamer implementation should be provided")
	}

	if handler == nil {
		panic("error: handler cannot be nil")
	}

	parentCtx, cancelFunc := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)

	if z.startFunc != nil {
		z.startFunc(parentCtx)
	}

	defer cancelFunc()

	var wg sync.WaitGroup
	wg.Add(len(streams))
	errChan := make(chan error, len(streams))
	for i, _ := range streams {
		go func(i int) {
			err := streams[i].Stream(parentCtx, handler)
			errChan <- err
			wg.Done()
		}(i)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	errors := make([]string, 0, len(streams))

	for err := range errChan {
		if err != nil {
			errors = append(errors, err.Error())
		}
	}

	if z.stopFunc != nil {
		z.stopFunc()
	}

	if len(errors) > 0 {
		return fmt.Errorf("stream run error: %s\n", strings.Join(errors, "\n"))
	}

	return fmt.Errorf("clean shutdown of streams")

}
