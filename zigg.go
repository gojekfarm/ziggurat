package ziggurat

import (
	"context"
	"errors"
	"github.com/sethvargo/go-signalcontext"
	"sync/atomic"
	"syscall"
)

type Ziggurat struct {
	handler    MessageHandler
	doneChan   chan error
	logLevel   string
	startFunc  StartFunction
	stopFunc   StopFunction
	isRunning  int32
	routes     Routes
	routeNames []string
	streams    Streams
}

func NewApp(opts ...ZigOptions) *Ziggurat {
	ziggurat := &Ziggurat{
		doneChan: make(chan error),
		streams:  New(),
	}
	for _, opts := range opts {
		opts(ziggurat)
	}
	if ziggurat.logLevel == "" {
		ziggurat.logLevel = "info"
	}
	return ziggurat
}

func (z *Ziggurat) appendRouteNames(routes Routes) {
	for name, _ := range routes {
		z.routeNames = append(z.routeNames, name)
	}
}

func (z *Ziggurat) Run(ctx context.Context, handler MessageHandler, routes Routes) chan error {
	if atomic.LoadInt32(&z.isRunning) == 1 {
		LogError(errors.New("attempted to call `Run` on an already running app"), "app run error", nil)
		return nil
	}

	z.appendRouteNames(routes)
	parentCtx, canceler := signalcontext.Wrap(ctx, syscall.SIGINT, syscall.SIGTERM)

	ConfigureLogger(z.logLevel)

	z.handler = handler
	z.routes = routes

	atomic.StoreInt32(&z.isRunning, 1)
	go func() {
		err := <-z.start(parentCtx, z.startFunc)
		LogError(err, "error starting streams", nil)
		canceler()
		atomic.StoreInt32(&z.isRunning, 0)
		z.stop(parentCtx, z.stopFunc)
		z.doneChan <- parentCtx.Err()
	}()
	return z.doneChan
}

func (z *Ziggurat) start(ctx context.Context, startCallback StartFunction) chan error {
	if startCallback != nil {
		LogInfo("ZIGGURAT: invoking start callback", nil)
		startCallback(ctx, z.routeNames)
	}

	LogInfo("ZIGGURAT: starting kafka streams", nil)

	streamsStop := z.streams.Consume(ctx, z.routes, z.handler)
	return streamsStop
}

func (z *Ziggurat) stop(ctx context.Context, stopFunc StopFunction) {
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

func (z *Ziggurat) OnStart(function StartFunction) {
	z.startFunc = function
}

func (z *Ziggurat) OnStop(function StopFunction) {
	z.stopFunc = function
}
