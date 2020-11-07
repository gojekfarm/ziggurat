package ziggurat

import (
	"context"
	"errors"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
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
	config          *Config
	router          StreamRouter
	metricPublisher MetricPublisher
	interruptChan   chan os.Signal
	doneChan        chan int
	startFunc       StartFunction
	stopFunc        StopFunction
	ctx             context.Context
	cancelFun       context.CancelFunc
	isRunning       int32
}

type RunOptions struct {
	HTTPConfigFunc func(a App, r *httprouter.Router)
	StartCallback  func(a App)
	StopCallback   func()
}

func NewApp() *Ziggurat {
	ctx, cancelFn := context.WithCancel(context.Background())
	commandLineOptions := ParseCommandLineArguments()
	parseConfig(commandLineOptions)
	log.Info().Msg("[ZIG APP] successfully parsed config")
	config := getConfig()
	if validationErr := config.validate(); validationErr != nil {
		log.Fatal().Err(validationErr).Msg("[ZIG APP] error creating app")
	}
	return &Ziggurat{
		ctx:           ctx,
		cancelFun:     cancelFn,
		config:        &config,
		stopFunc:      func() {},
		interruptChan: make(chan os.Signal),
		doneChan:      make(chan int),
	}
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
		z.metricPublisher = NewStatsD(z.config)
	}
	if z.messageRetry == nil {
		z.messageRetry = NewRabbitMQRetry(z.config)
	}

	if z.httpServer == nil {
		z.httpServer = NewDefaultHTTPServer(z.config)
	}
}

func (z *Ziggurat) configureHTTPRoutes(configFunc func(a App, r *httprouter.Router)) {
	z.httpServer.ConfigureHTTPRoutes(z, configFunc)
}

func (z *Ziggurat) Run(router StreamRouter, runOptions RunOptions) chan int {
	if atomic.LoadInt32(&z.isRunning) == 1 {
		log.Error().Err(errors.New("app is already running")).Msg("[ZIG APP]")
		return z.doneChan
	}

	{
		signal.Notify(z.interruptChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
		configureLogger(z.config.LogLevel)
		z.router = router
		z.configureDefaults()
		if runOptions.HTTPConfigFunc != nil {
			z.configureHTTPRoutes(runOptions.HTTPConfigFunc)
		}
		go func() {
			z.start(runOptions.StartCallback, runOptions.StopCallback)
			close(z.doneChan)
		}()
		atomic.StoreInt32(&z.isRunning, 1)
		return z.doneChan
	}
}

func (z *Ziggurat) start(startCallback StartFunction, stopCallback StopFunction) {

	if err := z.metricPublisher.Start(z); err != nil {
		log.Error().Err(err).Msg("[ZIG APP]")
	}

	z.httpServer.Start(z)

	if z.config.Retry.Enabled {
		_, err := z.messageRetry.Start(z)
		if err != nil {
			log.Error().Err(err).Msg("")
		}
	}

	routerStopChan, routerStartErr := z.router.Start(z)
	if routerStartErr != nil {
		log.Fatal().Err(routerStartErr)
	}

	if startCallback != nil {
		startCallback(z)
	}
	halt := func(routerStopChan chan int) {
		z.cancelFun()
		if routerStopChan != nil {
			<-routerStopChan
		}
		atomic.StoreInt32(&z.isRunning, 0)
		log.Info().Msg("[ZIG APP] router poll complete")
		z.stop(stopCallback)
	}
	// Wait for router poll to complete
	select {
	case <-routerStopChan:
		halt(nil)
	case <-z.interruptChan:
		log.Info().Msg("[ZIG APP] CTRL+C interrupt received")
		halt(routerStopChan)
	}
}

func (z *Ziggurat) Stop() {
	z.cancelFun()
	z.stop(z.stopFunc)
}

func (z *Ziggurat) stop(stopFunc StopFunction) {
	if err := z.messageRetry.Stop(); err != nil {
		log.Error().Err(err).Msg("[ZIG APP] error stopping retry")
	}

	if err := z.httpServer.Stop(); err != nil {
		log.Error().Err(err).Msg("[ZIG APP] error stopping http server")
	}

	if err := z.metricPublisher.Stop(); err != nil {
		log.Error().Err(err).Msg("[ZIG APP] error stopping metrics")
	}
	log.Info().Msg("[ZIG APP] invoking actor stop callback")

	if stopFunc != nil {
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
	return z.config
}
