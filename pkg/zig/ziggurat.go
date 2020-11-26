package zig

import (
	"context"
	"errors"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/cmdparser"
	"github.com/gojekfarm/ziggurat-go/pkg/kstream"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/metrics"
	"github.com/gojekfarm/ziggurat-go/pkg/retry"
	"github.com/gojekfarm/ziggurat-go/pkg/server"
	"github.com/gojekfarm/ziggurat-go/pkg/vconf"
	"github.com/gojekfarm/ziggurat-go/pkg/void"
	ztype "github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zerror"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

type Ziggurat struct {
	httpServer      ztype.HttpServer
	messageRetry    ztype.MessageRetry
	configStore     ztype.ConfigStore
	handler         ztype.MessageHandler
	metricPublisher ztype.MetricPublisher
	interruptChan   chan os.Signal
	doneChan        chan struct{}
	startFunc       ztype.StartFunction
	stopFunc        ztype.StopFunction
	ctx             context.Context
	cancelFun       context.CancelFunc
	isRunning       int32
	routes          []string
}

func NewApp() *Ziggurat {
	ctx, cancelFn := context.WithCancel(context.Background())
	ziggurat := &Ziggurat{
		ctx:           ctx,
		cancelFun:     cancelFn,
		configStore:   vconf.NewViperConfig(),
		stopFunc:      func() {},
		interruptChan: make(chan os.Signal),
		doneChan:      make(chan struct{}),
	}
	return ziggurat
}

func interruptHandler(c chan os.Signal, cancelFunc context.CancelFunc) {
	<-c
	cancelFunc()
}

func NewOpts() *ztype.RunOptions {
	return &ztype.RunOptions{
		HTTPConfigFunc:  func(a ztype.App, h http.Handler) {},
		StartCallback:   func(a ztype.App) {},
		StopCallback:    func() {},
		HTTPServer:      server.NewDefaultHTTPServer,
		Retry:           retry.NewRabbitMQRetry,
		MetricPublisher: metrics.NewStatsD,
	}
}

func (z *Ziggurat) loadConfig() error {
	commandLineOptions := cmdparser.ParseCommandLineArguments()
	z.configStore.Parse(commandLineOptions)
	logger.LogInfo("successfully parsed application config", nil)
	if validationErr := z.configStore.Validate(vconf.ConfigRules); validationErr != nil {
		return validationErr
	}
	logger.ConfigureLogger(z.configStore.Config().LogLevel)
	return nil
}

func (z *Ziggurat) setDefaultOpts(ro *ztype.RunOptions) {
	if ro.Retry == nil {
		ro.Retry = void.NewVoidRetry
	}
	if ro.MetricPublisher == nil {
		ro.MetricPublisher = void.NewVoidMetrics
	}
	if ro.HTTPServer == nil {
		ro.HTTPServer = void.NewVoidServer
	}

	if ro.HTTPConfigFunc == nil {
		ro.HTTPConfigFunc = func(a ztype.App, h http.Handler) {}
	}

	if ro.StopCallback == nil {
		ro.StopCallback = func() {}
	}

	if ro.StartCallback == nil {
		ro.StartCallback = func(a ztype.App) {}
	}

}

func (z *Ziggurat) Run(handler ztype.MessageHandler, routes []string, opts ...ztype.Opts) chan struct{} {
	if atomic.LoadInt32(&z.isRunning) == 1 {
		logger.LogError(errors.New("attempted to call `Run` on an already running app"), "app run error", nil)
		return nil
	}
	runOptions := NewOpts()
	if len(routes) < 1 {
		logger.LogFatal(zerror.ErrNoRoutesFound, "app run error", nil)
	}

	for _, opt := range opts {
		opt(runOptions)
	}

	z.handler = handler
	z.routes = routes

	signal.Notify(z.interruptChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)
	go interruptHandler(z.interruptChan, z.cancelFun)
	logger.LogFatal(z.loadConfig(), "ziggurat app load config", nil)

	z.setDefaultOpts(runOptions)
	z.messageRetry = runOptions.Retry(z.configStore)
	z.httpServer = runOptions.HTTPServer(z.configStore)
	z.metricPublisher = runOptions.MetricPublisher(z.configStore)
	z.httpServer.ConfigureHTTPRoutes(z, runOptions.HTTPConfigFunc)
	z.stopFunc = runOptions.StopCallback
	z.startFunc = runOptions.StartCallback

	atomic.StoreInt32(&z.isRunning, 1)
	go func() {
		z.start(runOptions.StartCallback)
		atomic.StoreInt32(&z.isRunning, 0)
		z.stop(z.stopFunc)
		close(z.doneChan)
	}()
	return z.doneChan
}

func (z *Ziggurat) start(startCallback ztype.StartFunction) {

	logger.LogError(z.metricPublisher.Start(z), "error starting metric publisher", nil)
	z.httpServer.Start(z)
	if z.configStore.Config().Retry.Enabled {
		err := z.messageRetry.Start(z)
		logger.LogFatal(err, "error starting retries", nil)
	}
	streamsStop, streamStartErr := kstream.NewKafkaStreams().Start(z)
	logger.LogFatal(streamStartErr, "ziggurat: router start error", nil)

	startCallback(z)

	halt := func(streamStop chan struct{}) {
		z.cancelFun()
		if streamStop != nil {
			<-streamStop
		}
		logger.LogInfo("stream poll complete", nil)
	}
	// Wait for router poll to complete
	select {
	case <-streamsStop:
		halt(nil)
	case <-z.interruptChan:
		logger.LogInfo("ziggurat app: CTRL+C interrupt received", nil)
		halt(streamsStop)
	}
}

func (z *Ziggurat) Stop() {
	z.cancelFun()
}

func (z *Ziggurat) stop(stopFunc ztype.StopFunction) {
	logger.LogError(z.messageRetry.Stop(), "failed to stop retries", nil)
	logger.LogError(z.httpServer.Stop(z), "failed to stop http server", nil)
	logger.LogError(z.metricPublisher.Stop(), "failed to stop metric publisher", nil)
	stopFunc()
}

func (z *Ziggurat) Context() context.Context {
	return z.ctx
}

func (z *Ziggurat) Routes() []string {
	return z.routes
}

func (z *Ziggurat) Handler() ztype.MessageHandler {
	return z.handler
}

func (z *Ziggurat) MessageRetry() ztype.MessageRetry {
	return z.messageRetry
}

func (z *Ziggurat) MetricPublisher() ztype.MetricPublisher {
	return z.metricPublisher
}

func (z *Ziggurat) HTTPServer() ztype.HttpServer {
	return z.httpServer
}

func (z *Ziggurat) Config() *basic.Config {
	return z.configStore.Config()
}

func (z *Ziggurat) IsRunning() bool {
	if atomic.LoadInt32(&z.isRunning) == 1 {
		return true
	}
	return false
}

func (z *Ziggurat) ConfigStore() ztype.ConfigStore {
	return z.configStore
}
