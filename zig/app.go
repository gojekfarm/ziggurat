package zig

import (
	"context"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
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
	log.Info().Msg(" successfully parsed appconf")
	if validationErr := z.appconf.Validate(); validationErr != nil {
		return validationErr
	}
	configureLogger(z.appconf.Config().LogLevel)
	return nil
}

func (z *Ziggurat) Run(router StreamRouter, runOptions RunOptions) chan struct{} {
	signal.Notify(z.interruptChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	logErrAndExit(z.loadConfig(), "")
	z.router = router
	z.configureDefaults()
	if runOptions.HTTPConfigFunc != nil {
		z.configureHTTPRoutes(runOptions.HTTPConfigFunc)
	}
	go func() {
		z.start(runOptions.StartCallback, runOptions.StopCallback)
		close(z.doneChan)
	}()
	return z.doneChan
}

func (z *Ziggurat) start(startCallback StartFunction, stopCallback StopFunction) {

	if err := z.metricPublisher.Start(z); err != nil {
		log.Error().Err(err).Msg("")
	}

	z.httpServer.Start(z)

	if z.appconf.Config().Retry.Enabled {
		_, err := z.messageRetry.Start(z)
		logError(err, "")
	}

	routerStopChan, routerStartErr := z.router.Start(z)
	logErrAndExit(routerStartErr, "")

	if startCallback != nil {
		startCallback(z)
	}
	halt := func(routerStopChan chan int) {
		z.cancelFun()
		if routerStopChan != nil {
			<-routerStopChan
		}
		log.Info().Msg(" router poll complete")
		z.stop(stopCallback)
	}
	// Wait for router poll to complete
	select {
	case <-routerStopChan:
		halt(nil)
	case <-z.interruptChan:
		log.Info().Msg(" CTRL+C interrupt received")
		halt(routerStopChan)
	}
}

func (z *Ziggurat) Stop() {
	z.cancelFun()
	z.stop(z.stopFunc)
}

func (z *Ziggurat) stop(stopFunc StopFunction) {
	logError(z.messageRetry.Stop(), "failed to stop retries")
	logError(z.httpServer.Stop(), "failed to stop http server")
	logError(z.metricPublisher.Stop(), "failed to stop metric publisher")

	if stopFunc != nil {
		logInfo("invoking actor stop callback")
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
