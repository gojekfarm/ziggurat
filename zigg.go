package ziggurat

import (
	"context"
	"github.com/sethvargo/go-signalcontext"
	"sync/atomic"
	"syscall"
)

type Ziggurat struct {
	Handler   Handler
	Logger    StructuredLogger
	startFunc StartFunction
	stopFunc  StopFunction
	isRunning int32
	streams   Streams
}

func NewApp(opts ...ZigOptions) *Ziggurat {
	ziggurat := &Ziggurat{}
	for _, opts := range opts {
		opts(ziggurat)
	}
	if ziggurat.Logger == nil {
		ziggurat.Logger = NewLogger("info")
	}
	return ziggurat
}

func (z *Ziggurat) Run(ctx context.Context, streams Streams, handler Handler) chan error {
	if atomic.LoadInt32(&z.isRunning) == 1 {
		return nil
	}

	if z.Logger == nil {
		z.Logger = NewLogger("info")
	}

	if streams == nil {
		panic("`streams` cannot be nil")
	}

	if handler == nil {
		panic("`handler` cannot be nil")
	}

	z.streams = streams

	doneChan := make(chan error)
	parentCtx, canceler := signalcontext.Wrap(ctx, syscall.SIGINT, syscall.SIGTERM)

	z.Handler = handler

	atomic.StoreInt32(&z.isRunning, 1)
	go func() {
		err := <-z.start(parentCtx, z.startFunc)
		z.Logger.Error("error starting streams", err)
		canceler()
		atomic.StoreInt32(&z.isRunning, 0)
		z.stop(z.stopFunc)
		doneChan <- parentCtx.Err()
	}()
	return doneChan
}

func (z *Ziggurat) start(ctx context.Context, startCallback StartFunction) chan error {
	if startCallback != nil {
		startCallback(ctx)
	}

	streamsStop := z.streams.Consume(ctx, z.Handler)
	return streamsStop
}

func (z *Ziggurat) stop(stopFunc StopFunction) {
	if stopFunc != nil {
		stopFunc()
	}
}

func (z *Ziggurat) IsRunning() bool {
	if atomic.LoadInt32(&z.isRunning) == 1 {
		return true
	}
	return false
}

func (z *Ziggurat) StartFunc(function StartFunction) {
	z.startFunc = function
}

func (z *Ziggurat) StopFunc(function StopFunction) {
	z.stopFunc = function
}
