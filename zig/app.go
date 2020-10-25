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
	Retrier         MessageRetrier
	MetricPublisher MetricPublisher
}

type App struct {
	httpServer      HttpServer
	retrier         MessageRetrier
	config          *Config
	router          StreamRouter
	metricPublisher MetricPublisher
	interruptChan   chan os.Signal
	doneChan        chan bool
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
	log.Info().Msg("successfully parsed config")
	config := getConfig()
	if validationErr := config.validate(); validationErr != nil {
		log.Fatal().Err(validationErr).Msg("error creating app")
	}
	return &App{
		ctx:           ctx,
		cancelFun:     cancelFn,
		config:        &config,
		stopFunc:      func() {},
		interruptChan: make(chan os.Signal),
		doneChan:      make(chan bool),
	}
}

func (a *App) Configure(configFunc func(app *App) Options) {
	options := configFunc(a)
	a.metricPublisher = options.MetricPublisher
	a.httpServer = options.HttpServer
	a.retrier = options.Retrier
	//configure defaults if components are nil
}

func (a *App) configureDefaults() {
	if a.metricPublisher == nil {
		a.metricPublisher = NewStatsD(a.config)
	}
	if a.retrier == nil {
		a.retrier = NewRabbitMQ(a.config)
	}

	if a.httpServer == nil {
		a.httpServer = NewDefaultHTTPServer(a.config)
	}
}

func (a *App) ConfigureHTTPRoutes(configFunc func(a *App, r *httprouter.Router)) {
	a.httpServer.ConfigureHTTPRoutes(a, configFunc)
}

func (a *App) Run(router StreamRouter, runOptions RunOptions) {
	signal.Notify(a.interruptChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	configureLogger(a.config.LogLevel)
	a.router = router
	a.configureDefaults()
	if runOptions.HTTPConfigFunc != nil {
		a.ConfigureHTTPRoutes(runOptions.HTTPConfigFunc)
	}
	a.start(runOptions.StartCallback, runOptions.StopCallback)
}

func (a *App) start(startCallback StartFunction, stopCallback StopFunction) {

	a.httpServer.Start(a)
	if err := a.metricPublisher.Start(a); err != nil {
		log.Error().Err(err).Msg("")
	}

	if a.config.Retry.Enabled {
		retrierStopChan, err := a.retrier.Start(a)

		go func() {
			<-retrierStopChan
			log.Error().Err(ErrRetryConsumerStopped).Msg("")
		}()
		if err != nil {
			log.Error().Err(err).Msg("")
		}
	}

	stopChan, routerStartErr := a.router.Start(a)
	if routerStartErr != nil {
		log.Fatal().Err(routerStartErr)
	}

	if startCallback != nil {
		startCallback(a)
	}
	halt := func() {
		a.cancelFun()
		a.stop(stopCallback)
	}
	// Wait for router poll to complete
	select {
	case <-stopChan:
		halt()
	case <-a.interruptChan:
		halt()
	}

}

func (a *App) stop(stopFunc StopFunction) {
	if err := a.retrier.Stop(); err != nil {
		log.Error().Err(err).Msg("error stopping retrier")
	}

	if err := a.httpServer.Stop(); err != nil {
		log.Error().Err(err).Msg("error stopping http server")
	}

	if err := a.metricPublisher.Stop(); err != nil {
		log.Error().Err(err).Msg("error stopping metrics")
	}
	log.Info().Msg("invoking actor stop function")

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

func (a *App) Retrier() MessageRetrier {
	return a.retrier
}

func (a *App) MetricPub() MetricPublisher {
	return a.metricPublisher
}

func (a *App) HTTPServer() HttpServer {
	return a.httpServer
}

func (a *App) Config() *Config {
	return a.config
}
