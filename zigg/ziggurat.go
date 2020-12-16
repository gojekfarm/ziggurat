package zigg

import (
	"context"
	"errors"
	"github.com/gojekfarm/ziggurat/kstream"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/zlog"
	"github.com/gojekfarm/ziggurat/ztype"
	"github.com/sethvargo/go-signalcontext"
	"sync/atomic"
)

type Ziggurat struct {
	handler   ztype.MessageHandler
	doneChan  chan struct{}
	logLevel  string
	startFunc ztype.StartFunction
	stopFunc  ztype.StopFunction
	ctx       context.Context
	cancelFun context.CancelFunc
	isRunning int32
	routes    zbase.Routes
	streams   ztype.Streams
}

func New(opts ...ZigOptions) *Ziggurat {
	ziggurat := &Ziggurat{
		doneChan: make(chan struct{}),
		streams:  kstream.New(),
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

func (z *Ziggurat) Run(handler ztype.MessageHandler, routes zbase.Routes) chan struct{} {
	if atomic.LoadInt32(&z.isRunning) == 1 {
		zlog.LogError(errors.New("attempted to call `Run` on an already running app"), "app run error", nil)
		return nil
	}

	zlog.ConfigureLogger(z.logLevel)

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

func (z *Ziggurat) start(startCallback ztype.StartFunction) chan struct{} {

	if startCallback != nil {
		zlog.LogInfo("ZIGGURAT: invoking start callback", nil)
		startCallback(z)
	}

	zlog.LogInfo("ZIGGURAT: starting kafka streams", nil)

	streamsStop, streamsStartErr := z.streams.Start(z)
	if streamsStartErr != nil {
		zlog.LogFatal(streamsStartErr, "error starting kafka streams", nil)
	}

	return streamsStop

}

func (z *Ziggurat) Stop() {
	zlog.LogInfo("stopping app: cancelling context", nil)
	z.cancelFun()
	zlog.LogInfo("stopping app: stopping streams", nil)
	z.streams.Stop()
	z.stop(z.stopFunc)
}

func (z *Ziggurat) stop(stopFunc ztype.StopFunction) {
	if stopFunc != nil {
		stopFunc(z)
	}
}

func (z *Ziggurat) Context() context.Context {
	return z.ctx
}

func (z *Ziggurat) Routes() zbase.Routes {
	return z.routes
}

func (z *Ziggurat) Handler() ztype.MessageHandler {
	return z.handler
}

func (z *Ziggurat) IsRunning() bool {
	if atomic.LoadInt32(&z.isRunning) == 1 {
		return true
	}
	return false
}

func (z *Ziggurat) OnStart(function ztype.StartFunction) {
	z.startFunc = function
}

func (z *Ziggurat) OnStop(function ztype.StopFunction) {
	z.stopFunc = function
}
