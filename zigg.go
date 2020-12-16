package ziggurat

import (
	"context"
	"errors"
	"github.com/sethvargo/go-signalcontext"
	"sync/atomic"
)

type Ziggurat struct {
	handler   MessageHandler
	doneChan  chan struct{}
	logLevel  string
	startFunc StartFunction
	stopFunc  StopFunction
	ctx       context.Context
	cancelFun context.CancelFunc
	isRunning int32
	routes    Routes
	streams   Streams
}

func NewApp(opts ...ZigOptions) *Ziggurat {
	ziggurat := &Ziggurat{
		doneChan: make(chan struct{}),
		streams:  New(),
	}
	for _, opts := range opts {
		opts(ziggurat)
	}
	if ziggurat.logLevel == "" {
		ziggurat.logLevel = "info"
	}
	if ziggurat.ctx == nil {
		ctx, cancelFn := signalcontext.OnInterrupt()
		ziggurat.ctx = ctx
		ziggurat.cancelFun = cancelFn
	}
	return ziggurat
}

func (z *Ziggurat) Run(handler MessageHandler, routes Routes) chan struct{} {
	if atomic.LoadInt32(&z.isRunning) == 1 {
		LogError(errors.New("attempted to call `Run` on an already running app"), "app run error", nil)
		return nil
	}

	ConfigureLogger(z.logLevel)

	z.handler = handler
	z.routes = routes

	atomic.StoreInt32(&z.isRunning, 1)
	go func() {
		<-z.start(z.startFunc)
		z.cancelFun()
		atomic.StoreInt32(&z.isRunning, 0)
		z.stop(z.stopFunc)
		close(z.doneChan)
	}()
	return z.doneChan
}

func (z *Ziggurat) start(startCallback StartFunction) chan struct{} {

	if startCallback != nil {
		LogInfo("ZIGGURAT: invoking start callback", nil)
		startCallback(z)
	}

	LogInfo("ZIGGURAT: starting kafka streams", nil)

	streamsStop, streamsStartErr := z.streams.Start(z)
	if streamsStartErr != nil {
		LogFatal(streamsStartErr, "error starting kafka streams", nil)
	}

	return streamsStop

}

func (z *Ziggurat) Stop() {
	LogInfo("stopping app: cancelling context", nil)
	z.cancelFun()
	LogInfo("stopping app: stopping streams", nil)
	z.streams.Stop()
	z.stop(z.stopFunc)
}

func (z *Ziggurat) stop(stopFunc StopFunction) {
	if stopFunc != nil {
		stopFunc(z)
	}
}

func (z *Ziggurat) Context() context.Context {
	return z.ctx
}

func (z *Ziggurat) Routes() Routes {
	return z.routes
}

func (z *Ziggurat) Handler() MessageHandler {
	return z.handler
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
