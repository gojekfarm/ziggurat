package zig

import (
	"context"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

type Options struct {
	HttpServer      HttpServer
	Retrier         MessageRetry
	MetricPublisher MetricPublisher
}

type App struct {
	httpServer      HttpServer
	messageRetry    MessageRetry
	config          *Config
	router          StreamRouter
	metricPublisher MetricPublisher
	interruptChan   chan os.Signal
	doneChan        chan int
	startFunc       StartFunction
	stopFunc        StopFunction
	ctx             context.Context
	cancelFun       context.CancelFunc
}

type RunOptions struct {
	HTTPConfigFunc func(a *App, r *httprouter.Router)
	StartCallback  func(a *App)
	StopCallback   func()
}

func NewApp() *App {
	ctx, cancelFn := context.WithCancel(context.Background())
	commandLineOptions := ParseCommandLineArguments()
	parseConfig(commandLineOptions)
	log.Info().Msg("[ZIG APP] successfully parsed config")
	config := getConfig()
	if validationErr := config.validate(); validationErr != nil {
		log.Fatal().Err(validationErr).Msg("[ZIG APP] error creating app")
	}
	return &App{
		ctx:           ctx,
		cancelFun:     cancelFn,
		config:        &config,
		stopFunc:      func() {},
		interruptChan: make(chan os.Signal),
		doneChan:      make(chan int),
	}
}

func (a *App) Configure(configFunc func(app *App) Options) {
	options := configFunc(a)
	a.metricPublisher = options.MetricPublisher
	a.httpServer = options.HttpServer
	a.messageRetry = options.Retrier
	//configure defaults if components are nil
}

func (a *App) configureDefaults() {
	if a.metricPublisher == nil {
		a.metricPublisher = NewStatsD(a.config)
	}
	if a.messageRetry == nil {
		a.messageRetry = NewRabbitMQRetry(a.config)
	}

	if a.httpServer == nil {
		a.httpServer = NewDefaultHTTPServer(a.config)
	}
}

func (a *App) configureHTTPRoutes(configFunc func(a *App, r *httprouter.Router)) {
	a.httpServer.ConfigureHTTPRoutes(a, configFunc)
}

func (a *App) Run(router StreamRouter, runOptions RunOptions) chan int {

	signal.Notify(a.interruptChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	configureLogger(a.config.LogLevel)
	a.router = router
	a.configureDefaults()
	if runOptions.HTTPConfigFunc != nil {
		a.configureHTTPRoutes(runOptions.HTTPConfigFunc)
	}
	go func() {
		a.start(runOptions.StartCallback, runOptions.StopCallback)
		close(a.doneChan)
	}()
	return a.doneChan
}

func (a *App) start(startCallback StartFunction, stopCallback StopFunction) {

	if err := a.metricPublisher.Start(a); err != nil {
		log.Error().Err(err).Msg("[ZIG APP]")
	}

	a.httpServer.Start(a)

	if a.config.Retry.Enabled {
		retryStopChan, err := a.messageRetry.Start(a)

		go func() {
			<-retryStopChan
			log.Error().Err(ErrRetryConsumerStopped).Msg("")
		}()
		if err != nil {
			log.Error().Err(err).Msg("")
		}
	}

	routerStopChan, routerStartErr := a.router.Start(a)
	if routerStartErr != nil {
		log.Fatal().Err(routerStartErr)
	}

	if startCallback != nil {
		startCallback(a)
	}
	halt := func(routerStopChan chan int) {
		a.cancelFun()
		if routerStopChan != nil {
			<-routerStopChan
		}
		log.Info().Msg("[ZIG APP] router poll complete")
		a.stop(stopCallback)
	}
	// Wait for router poll to complete
	select {
	case <-routerStopChan:
		halt(nil)
	case <-a.interruptChan:
		log.Info().Msg("[ZIG APP] CTRL+C interrupt received")
		halt(routerStopChan)
	}
}

func (a *App) stop(stopFunc StopFunction) {
	if err := a.messageRetry.Stop(); err != nil {
		log.Error().Err(err).Msg("[ZIG APP] error stopping retry")
	}

	if err := a.httpServer.Stop(); err != nil {
		log.Error().Err(err).Msg("[ZIG APP] error stopping http server")
	}

	if err := a.metricPublisher.Stop(); err != nil {
		log.Error().Err(err).Msg("[ZIG APP] error stopping metrics")
	}
	log.Info().Msg("[ZIG APP] invoking actor stop callback")

	if stopFunc != nil {
		stopFunc()
	}
}

func (a *App) Context() context.Context {
	return a.ctx
}

func (a *App) Router() StreamRouter {
	return a.router
}

func (a *App) MessageRetry() MessageRetry {
	return a.messageRetry
}

func (a *App) MetricPublisher() MetricPublisher {
	return a.metricPublisher
}

func (a *App) HTTPServer() HttpServer {
	return a.httpServer
}

func (a *App) Config() *Config {
	return a.config
}
