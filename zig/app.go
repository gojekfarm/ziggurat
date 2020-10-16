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
	StopFunc          StopFunction
	HttpConfigureFunc func(r *httprouter.Router)
}

type App struct {
	HttpServer      HttpServer
	Retrier         MessageRetrier
	Config          *Config
	StreamRouter    *StreamRouter
	MetricPublisher MetricPublisher
	interruptChan   chan os.Signal
	doneChan        chan bool
	startFunc       StartFunction
	stopFunc        StopFunction
	ctx             context.Context
	cancelFun       context.CancelFunc
}

func NewApp() *App {
	ctx, cancelFn := context.WithCancel(context.Background())
	commandLineOptions := parseCommandLineArguments()
	parseConfig(commandLineOptions)
	log.Info().Msg("successfully parsed config")
	config := getConfig()
	if validationErr := config.validate(); validationErr != nil {
		log.Fatal().Err(validationErr).Msg("error creating app")
	}
	return &App{
		ctx:           ctx,
		cancelFun:     cancelFn,
		Config:        &config,
		stopFunc:      func() {},
		interruptChan: make(chan os.Signal),
		doneChan:      make(chan bool),
	}
}

func (a *App) Configure(options Options) {
	a.MetricPublisher = options.MetricPublisher
	a.HttpServer = options.HttpServer
	a.Retrier = options.Retrier
	a.stopFunc = options.StopFunc
	if options.HttpConfigureFunc != nil {
		a.HttpServer.attachRoute(options.HttpConfigureFunc)
	}
	//configure defaults if components are nil
}

func (a *App) configureDefaults() {
	if a.MetricPublisher == nil {
		a.MetricPublisher = &StatsD{}
	}
	if a.Retrier == nil {
		a.Retrier = &RabbitRetrier{}
	}

	if a.HttpServer == nil {
		a.HttpServer = &DefaultHttpServer{}
	}
}

func (a *App) Run(router *StreamRouter, startCallback StartFunction) {
	signal.Notify(a.interruptChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	configureLogger(a.Config.LogLevel)
	a.StreamRouter = router
	a.configureDefaults()
	a.start(startCallback)
}

func (a *App) start(startCallback StartFunction) {
	if stopChan, routerStartErr := a.StreamRouter.Start(a); routerStartErr != nil {
		log.Fatal().Err(routerStartErr)
	} else {
		a.HttpServer.Start(a)
		if err := a.MetricPublisher.Start(a); err != nil {
			log.Error().Err(err).Msg("")
		}

		retrierStopChan, err := a.Retrier.Start(a)

		go func() {
			<-retrierStopChan
			log.Error().Err(ErrRetryConsumerStopped).Msg("")
		}()

		if err != nil {
			log.Error().Err(err).Msg("")
		}

		startCallback(a)
		halt := func() {
			a.cancelFun()
			a.stop()
		}
		// Wait for router poll to complete
		select {
		case <-stopChan:
			halt()
		case <-a.interruptChan:
			halt()
		}
	}

}

func (a *App) stop() {
	if err := a.Retrier.Stop(); err != nil {
		log.Error().Err(err).Msg("error stopping retrier")
	}

	if err := a.HttpServer.Stop(); err != nil {
		log.Error().Err(err).Msg("error stopping http server")
	}

	if err := a.MetricPublisher.Stop(); err != nil {
		log.Error().Err(err).Msg("error stopping metrics")
	}
	log.Info().Msg("invoking actor stop function")
	a.stopFunc()
}

func (a *App) Context() context.Context {
	return a.ctx
}
