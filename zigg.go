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
	logger     LeveledLogger
	doneChan   chan error
	logLevel   string
	startFunc  StartFunction
	stopFunc   StopFunction
	isRunning  int32
	routes     Routes
	routeNames []string
}

func NewApp(opts ...ZigOptions) *Ziggurat {
	ziggurat := &Ziggurat{
		doneChan: make(chan error),
	}
	for _, opts := range opts {
		opts(ziggurat)
	}
	if ziggurat.logger == nil {
		ziggurat.logger = NewLogger("info")
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
		return nil
	}

	if len(routes) < 1 {
		z.logger.Fatalf("error starting app: %v", errors.New("no routes found"))
	}

	z.appendRouteNames(routes)
	parentCtx, canceler := signalcontext.Wrap(ctx, syscall.SIGINT, syscall.SIGTERM)

	z.handler = handler
	z.routes = routes

	atomic.StoreInt32(&z.isRunning, 1)
	go func() {
		err := <-z.start(parentCtx, z.startFunc)
		z.logger.Errorf("error starting streams: %v", err)
		canceler()
		atomic.StoreInt32(&z.isRunning, 0)
		z.stop(z.stopFunc)
		z.doneChan <- parentCtx.Err()
	}()
	return z.doneChan
}

func (z *Ziggurat) start(ctx context.Context, startCallback StartFunction) chan error {
	if startCallback != nil {
		startCallback(ctx, z.routeNames)
	}

	streams := NewKafkaStreams(z.logger)
	streamsStop := streams.Consume(ctx, z.routes, z.handler)
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

func (z *Ziggurat) OnStart(function StartFunction) {
	z.startFunc = function
}

func (z *Ziggurat) OnStop(function StopFunction) {
	z.stopFunc = function
}
