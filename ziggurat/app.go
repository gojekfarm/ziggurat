package ziggurat

import (
	"context"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

type Options struct {
	HttpServer      HttpServer
	Retrier         MessageRetrier
	MetricPublisher MetricPublisher
	StopFunc        StopFunction
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

// Run method blocks the app and returns a channel to notify app stop
func (a *App) Run(router *StreamRouter, startCallback StartFunction) chan bool {
	signal.Notify(a.interruptChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	configureLogger(a.Config.LogLevel)
	a.StreamRouter = router
	a.configureDefaults()
	a.start(startCallback)
	a.doneChan <- true
	return a.doneChan
}

func (a *App) start(startCallback StartFunction) {
	if stopChan, routerStartErr := a.StreamRouter.Start(a); routerStartErr != nil {
		log.Fatal().Err(routerStartErr)
	} else {
		a.HttpServer.Start(a)
		if err := a.MetricPublisher.Start(a); err != nil {
			log.Error().Err(err).Msg("")
		}
		if err := a.Retrier.Start(a); err != nil {
			log.Error().Err(err).Msg("")
		}
		startCallback(a)
		// Wait for router poll to complete
		for {
			select {
			case <-stopChan:
				a.cancelFun()
				a.stop()
				return
			case <-a.interruptChan:
				a.cancelFun()
				a.stop()
				return
			}
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
