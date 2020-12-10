package zapp

import (
	"context"
	"errors"
	"github.com/gojekfarm/ziggurat/kstream"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/zlog"
	ztype "github.com/gojekfarm/ziggurat/ztype"
	"github.com/gojekfarm/ziggurat/zvoid"
	"github.com/sethvargo/go-signalcontext"
	"sync/atomic"
)

type Ziggurat struct {
	httpServer      ztype.Server
	messageRetry    ztype.MessageRetry
	configValidator ztype.ConfigValidator
	handler         ztype.MessageHandler
	metricPublisher ztype.MetricPublisher
	doneChan        chan struct{}
	logLevel        string
	startFunc       ztype.StartFunction
	stopFunc        ztype.StopFunction
	ctx             context.Context
	cancelFun       context.CancelFunc
	isRunning       int32
	routes          []zbase.StreamConfig
	streams         ztype.Streams
}

func New(opts ...ZigOptions) *Ziggurat {
	ctx, cancelFn := signalcontext.OnInterrupt()
	ziggurat := &Ziggurat{
		ctx:       ctx,
		cancelFun: cancelFn,
		doneChan:  make(chan struct{}),
		streams:   kstream.New(),
	}
	for _, opts := range opts {
		opts(ziggurat)
	}
	if ziggurat.logLevel == "" {
		ziggurat.logLevel = "info"
	}
	return ziggurat
}

func (z *Ziggurat) Run(handler ztype.MessageHandler, routes ...zbase.StreamConfig) chan struct{} {
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
	components := z.components()
	for i, _ := range components {
		c := components[i]
		if c != nil {
			switch t := c.(type) {
			case ztype.MessageRetry:
				zlog.LogFatal(c.Start(z), "error starting component", map[string]interface{}{"COMPONENT": "RETRIES"})
			default:
				zlog.LogError(c.Start(z), "error starting component", map[string]interface{}{"COMPONENT": t})
			}
		}
	}

	streamsStop, streamsStartErr := z.streams.Start(z)
	if streamsStartErr != nil {
		zlog.LogFatal(streamsStartErr, "error starting kafka streams", nil)
	}

	if startCallback != nil {
		startCallback(z)
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
	components := z.components()
	for i, _ := range components {
		c := components[i]
		if c != nil {
			c.Stop(z)
		}
	}
	if stopFunc != nil {
		stopFunc()
	}
}

func (z *Ziggurat) Context() context.Context {
	return z.ctx
}

func (z *Ziggurat) Routes() []zbase.StreamConfig {
	return z.routes
}

func (z *Ziggurat) Handler() ztype.MessageHandler {
	return z.handler
}

func (z *Ziggurat) MessageRetry() ztype.MessageRetry {
	if z.messageRetry == nil {
		return zvoid.NewRetry()
	}
	return z.messageRetry
}

func (z *Ziggurat) MetricPublisher() ztype.MetricPublisher {
	if z.metricPublisher == nil {
		return zvoid.NewMetrics()
	}
	return z.metricPublisher
}

func (z *Ziggurat) HTTPServer() ztype.Server {
	if z.httpServer == nil {
		return zvoid.NewServer()
	}
	return z.httpServer
}

func (z *Ziggurat) components() []ztype.StartStopper {
	return []ztype.StartStopper{z.metricPublisher, z.messageRetry, z.httpServer}
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
