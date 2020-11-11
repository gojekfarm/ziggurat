package zig

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

type Options struct {
	HttpServer      HttpServer
	Retrier         MessageRetry
	MetricPublisher MetricPublisher
}

type Ziggurat struct {
	httpServer      HttpServer
	messageRetry    MessageRetry
	appconf         ConfigReader
	router          StreamRouter
	metricPublisher MetricPublisher
	interruptChan   chan os.Signal
	doneChan        chan struct{}
	startFunc       StartFunction
	stopFunc        StopFunction
	ctx             context.Context
	cancelFun       context.CancelFunc
	isRunning       int32
}

type RunOptions struct {
	HTTPConfigFunc func(a App, h http.Handler)
	StartCallback  func(a App)
	StopCallback   func()
}

func NewApp() *Ziggurat {
	ctx, cancelFn := context.WithCancel(context.Background())
	z := &Ziggurat{
		ctx:           ctx,
		cancelFun:     cancelFn,
		appconf:       NewViperConfig(),
		stopFunc:      func() {},
		interruptChan: make(chan os.Signal),
		doneChan:      make(chan struct{}),
	}
	return z
}

func (z *Ziggurat) Configure(configFunc func(app App) Options) {
	options := configFunc(z)
	z.metricPublisher = options.MetricPublisher
	z.httpServer = options.HttpServer
	z.messageRetry = options.Retrier
	//configure defaults if components are nil
}

func (z *Ziggurat) configureDefaults() {
	if z.metricPublisher == nil {
		z.metricPublisher = NewStatsD(z.appconf)
	}
	if z.messageRetry == nil {
		z.messageRetry = NewRabbitMQRetry(z.appconf)
	}

	if z.httpServer == nil {
		z.httpServer = NewDefaultHTTPServer(z.appconf)
	}
}

func (z *Ziggurat) configureHTTPRoutes(configFunc func(a App, h http.Handler)) {
	z.httpServer.ConfigureHTTPRoutes(z, configFunc)
}

func (z *Ziggurat) loadConfig() error {
	commandLineOptions := ParseCommandLineArguments()
	z.appconf.Parse(commandLineOptions)
	logInfo("successfully parsed application config", nil)
	if validationErr := z.appconf.Validate(configRules); validationErr != nil {
		return validationErr
	}
	configureLogger(z.appconf.Config().LogLevel)
	return nil
}

func (z *Ziggurat) Run(router StreamRouter, runOptions RunOptions) chan struct{} {
	if atomic.LoadInt32(&z.isRunning) == 1 {
		logError(errors.New("attempted to call `Run` on an already running app"), "ziggurat app run", nil)
		return nil
	}
	signal.Notify(z.interruptChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	logFatal(z.loadConfig(), "ziggurat app load config", nil)
	z.router = router
	z.configureDefaults()
	if runOptions.HTTPConfigFunc != nil {
		z.configureHTTPRoutes(runOptions.HTTPConfigFunc)
	}
	atomic.StoreInt32(&z.isRunning, 1)
	go func() {
		z.start(runOptions.StartCallback, runOptions.StopCallback)
		atomic.StoreInt32(&z.isRunning, 0)
		z.stop(z.stopFunc)
		close(z.doneChan)
	}()
	return z.doneChan
}

func (z *Ziggurat) start(startCallback StartFunction, stopCallback StopFunction) {
	logError(z.metricPublisher.Start(z), "", nil)

	z.httpServer.Start(z)

	if z.appconf.Config().Retry.Enabled {
		_, err := z.messageRetry.Start(z)
		logError(err, "", nil)
	}

	routerStopChan, routerStartErr := z.router.Start(z)
	logFatal(routerStartErr, "", nil)

	if startCallback != nil {
		startCallback(z)
	}
	halt := func(routerStopChan chan int) {
		z.cancelFun()
		if routerStopChan != nil {
			<-routerStopChan
		}
		logInfo("ziggurat app: router poll completed without errors", nil)
	}
	// Wait for router poll to complete
	select {
	case <-routerStopChan:
		halt(nil)
	case <-z.interruptChan:
		logInfo("ziggurat app: CTRL+C interrupt received", nil)
		halt(routerStopChan)
	}
}

func (z *Ziggurat) Stop() {
	z.cancelFun()
}

func (z *Ziggurat) stop(stopFunc StopFunction) {
	logError(z.messageRetry.Stop(), "failed to stop retries", nil)
	logError(z.httpServer.Stop(z.ctx), "failed to stop http server", nil)
	logError(z.metricPublisher.Stop(), "failed to stop metric publisher", nil)

	if stopFunc != nil {
		logInfo("invoking actor stop callback", nil)
		stopFunc()
	}
}

func (z *Ziggurat) Context() context.Context {
	return z.ctx
}

func (z *Ziggurat) Router() StreamRouter {
	return z.router
}

func (z *Ziggurat) MessageRetry() MessageRetry {
	return z.messageRetry
}

func (z *Ziggurat) MetricPublisher() MetricPublisher {
	return z.metricPublisher
}

func (z *Ziggurat) HTTPServer() HttpServer {
	return z.httpServer
}

func (z *Ziggurat) Config() *Config {
	return z.appconf.Config()
}

func (z *Ziggurat) IsRunning() bool {
	if atomic.LoadInt32(&z.isRunning) == 1 {
		return true
	}
	return false
}
