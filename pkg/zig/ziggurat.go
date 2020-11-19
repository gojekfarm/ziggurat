package zig

import (
	"context"
	"errors"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/cmdparser"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/metrics"
	"github.com/gojekfarm/ziggurat-go/pkg/retry"
	"github.com/gojekfarm/ziggurat-go/pkg/server"
	"github.com/gojekfarm/ziggurat-go/pkg/vconf"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

type Ziggurat struct {
	httpServer      z.HttpServer
	messageRetry    z.MessageRetry
	cfgReader       z.ConfigReader
	router          z.StreamRouter
	metricPublisher z.MetricPublisher
	interruptChan   chan os.Signal
	doneChan        chan struct{}
	startFunc       z.StartFunction
	stopFunc        z.StopFunction
	ctx             context.Context
	cancelFun       context.CancelFunc
	isRunning       int32
}

func NewApp() *Ziggurat {
	ctx, cancelFn := context.WithCancel(context.Background())
	ziggurat := &Ziggurat{
		ctx:           ctx,
		cancelFun:     cancelFn,
		cfgReader:     vconf.NewViperConfig(),
		stopFunc:      func() {},
		interruptChan: make(chan os.Signal),
		doneChan:      make(chan struct{}),
	}
	return ziggurat
}

func (z *Ziggurat) Configure(configFunc func(app z.App) z.Options) {
	options := configFunc(z)
	z.metricPublisher = options.MetricPublisher
	z.httpServer = options.HttpServer
	z.messageRetry = options.Retry
	//configure defaults if components are nil
}

func (z *Ziggurat) configureDefaults() {
	if z.metricPublisher == nil {
		z.metricPublisher = metrics.NewStatsD(z.cfgReader)
	}
	if z.messageRetry == nil {
		z.messageRetry = retry.NewRabbitMQRetry(z.cfgReader)
	}

	if z.httpServer == nil {
		z.httpServer = server.NewDefaultHTTPServer(z.cfgReader)
	}
}

func (z *Ziggurat) configureHTTPRoutes(configFunc func(a z.App, h http.Handler)) {
	z.httpServer.ConfigureHTTPRoutes(z, configFunc)
}

func (z *Ziggurat) loadConfig() error {
	commandLineOptions := cmdparser.ParseCommandLineArguments()
	z.cfgReader.Parse(commandLineOptions)
	logger.LogInfo("successfully parsed application config", nil)
	if validationErr := z.cfgReader.Validate(vconf.ConfigRules); validationErr != nil {
		return validationErr
	}
	logger.ConfigureLogger(z.cfgReader.Config().LogLevel)
	return nil
}

func (z *Ziggurat) Run(router z.StreamRouter, runOptions z.RunOptions) chan struct{} {
	if atomic.LoadInt32(&z.isRunning) == 1 {
		logger.LogError(errors.New("attempted to call `Run` on an already running app"), "ziggurat app run", nil)
		return nil
	}
	signal.Notify(z.interruptChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)
	go interruptHandler(z.interruptChan, z.cancelFun)
	logger.LogFatal(z.loadConfig(), "ziggurat app load config", nil)
	z.router = router
	z.configureDefaults()
	if runOptions.HTTPConfigFunc != nil {
		z.configureHTTPRoutes(runOptions.HTTPConfigFunc)
	}
	atomic.StoreInt32(&z.isRunning, 1)
	go func() {
		z.start(runOptions.StartCallback)
		atomic.StoreInt32(&z.isRunning, 0)
		z.stop(z.stopFunc)
		close(z.doneChan)
	}()
	return z.doneChan
}

func (z *Ziggurat) start(startCallback z.StartFunction) {
	logger.LogError(z.metricPublisher.Start(z), "", nil)

	z.httpServer.Start(z)

	if z.cfgReader.Config().Retry.Enabled {
		logger.LogInfo("ziggurat: starting retry component", nil)
		err := z.messageRetry.Start(z)
		logger.LogFatal(err, "ziggurat: error starting retries", nil)
	}

	routerStopChan, routerStartErr := z.router.Start(z)
	logger.LogFatal(routerStartErr, "ziggurat: router start error", nil)

	if startCallback != nil {
		startCallback(z)
	}
	halt := func(routerStopChan chan struct{}) {
		z.cancelFun()
		if routerStopChan != nil {
			<-routerStopChan
		}
		logger.LogInfo("ziggurat app: router poll completed without errors", nil)
	}
	// Wait for router poll to complete
	select {
	case <-routerStopChan:
		halt(nil)
	case <-z.interruptChan:
		logger.LogInfo("ziggurat app: CTRL+C interrupt received", nil)
		halt(routerStopChan)
	}
}

func (z *Ziggurat) Stop() {
	z.cancelFun()
}

func (z *Ziggurat) stop(stopFunc z.StopFunction) {
	logger.LogError(z.messageRetry.Stop(), "failed to stop retries", nil)
	logger.LogError(z.httpServer.Stop(z.ctx), "failed to stop http server", nil)
	logger.LogError(z.metricPublisher.Stop(), "failed to stop metric publisher", nil)

	if stopFunc != nil {
		logger.LogInfo("invoking actor stop callback", nil)
		stopFunc()
	}
}

func (z *Ziggurat) Context() context.Context {
	return z.ctx
}

func (z *Ziggurat) Router() z.StreamRouter {
	return z.router
}

func (z *Ziggurat) MessageRetry() z.MessageRetry {
	return z.messageRetry
}

func (z *Ziggurat) MetricPublisher() z.MetricPublisher {
	return z.metricPublisher
}

func (z *Ziggurat) HTTPServer() z.HttpServer {
	return z.httpServer
}

func (z *Ziggurat) Config() *basic.Config {
	return z.cfgReader.Config()
}

func (z *Ziggurat) IsRunning() bool {
	if atomic.LoadInt32(&z.isRunning) == 1 {
		return true
	}
	return false
}

func (z *Ziggurat) ConfigReader() z.ConfigReader {
	return z.cfgReader
}

func interruptHandler(c chan os.Signal, cancelFunc context.CancelFunc) {
	<-c
	cancelFunc()
}
