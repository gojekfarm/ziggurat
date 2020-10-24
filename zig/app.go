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
	HttpServer        HttpServer
	Retrier           MessageRetrier
	MetricPublisher   MetricPublisher
	HTTPConfigureFunc func(r *httprouter.Router)
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

func (a *App) Configure(options Options) {
	a.metricPublisher = options.MetricPublisher
	a.httpServer = options.HttpServer
	a.retrier = options.Retrier
	if options.HTTPConfigureFunc != nil {
		a.httpServer.DefineRoutes(options.HTTPConfigureFunc)
	}
	//configure defaults if components are nil
}

func (a *App) configureDefaults() {
	if a.metricPublisher == nil {
		a.metricPublisher = &StatsD{}
	}
	if a.retrier == nil {
		a.retrier = &RabbitRetrier{}
	}

	if a.httpServer == nil {
		a.httpServer = &DefaultHttpServer{}
	}
}

func (a *App) Run(router StreamRouter, startCallback StartFunction, stopCallback StopFunction) {
	signal.Notify(a.interruptChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	configureLogger(a.config.LogLevel)
	a.router = router
	a.configureDefaults()
	a.start(startCallback, stopCallback)
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
