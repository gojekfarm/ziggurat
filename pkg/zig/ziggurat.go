package zig

import (
	"context"
	"errors"
	"github.com/gojekfarm/ziggurat-go/pkg/cmdparser"
	"github.com/gojekfarm/ziggurat-go/pkg/kstream"
	"github.com/gojekfarm/ziggurat-go/pkg/metrics"
	"github.com/gojekfarm/ziggurat-go/pkg/rabbitmq"
	"github.com/gojekfarm/ziggurat-go/pkg/rules"
	"github.com/gojekfarm/ziggurat-go/pkg/server"
	ztype "github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zconf"
	"github.com/gojekfarm/ziggurat-go/pkg/zerror"
	"github.com/gojekfarm/ziggurat-go/pkg/zlogger"
	"github.com/sethvargo/go-signalcontext"
	"net/http"
	"sync/atomic"
)

type Ziggurat struct {
	httpServer      ztype.Server
	messageRetry    ztype.MessageRetry
	configStore     ztype.ConfigStore
	configValidator ztype.ConfigValidator
	handler         ztype.MessageHandler
	metricPublisher ztype.MetricPublisher
	doneChan        chan struct{}
	startFunc       ztype.StartFunction
	stopFunc        ztype.StopFunction
	ctx             context.Context
	cancelFun       context.CancelFunc
	isRunning       int32
	routes          []string
}

func NewApp() *Ziggurat {
	ctx, cancelFn := signalcontext.OnInterrupt()
	ziggurat := &Ziggurat{
		ctx:             ctx,
		cancelFun:       cancelFn,
		configStore:     zconf.NewViperConfig(),
		configValidator: zconf.NewDefaultValidator(rules.DefaultRules),
		doneChan:        make(chan struct{}),
	}
	return ziggurat
}

func NewOpts() *RunOptions {
	return &RunOptions{
		HTTPConfigFunc:  func(a ztype.App, h http.Handler) {},
		StartCallback:   func(a ztype.App) {},
		StopCallback:    func() {},
		HTTPServer:      server.NewDefaultHTTPServer,
		Retry:           rabbitmq.NewRabbitMQRetry,
		MetricPublisher: metrics.NewStatsD,
	}
}

func (z *Ziggurat) loadConfig() error {
	commandLineOptions := cmdparser.ParseCommandLineArguments()
	if parseErr := z.configStore.Parse(commandLineOptions); parseErr != nil {
		return parseErr
	}
	if validatorErr := z.configValidator.Validate(z.ConfigStore().Config()); validatorErr != nil {
		return validatorErr
	}
	zlogger.LogInfo("successfully parsed application config", nil)
	zlogger.ConfigureLogger(z.configStore.Config().LogLevel)
	return nil
}

func (z *Ziggurat) Run(handler ztype.MessageHandler, routes []string, opts ...Opts) chan struct{} {
	if atomic.LoadInt32(&z.isRunning) == 1 {
		zlogger.LogError(errors.New("attempted to call `Run` on an already running app"), "app run error", nil)
		return nil
	}
	runOptions := NewOpts()
	if len(routes) < 1 {
		zlogger.LogFatal(zerror.ErrNoRoutesFound, "app run error", nil)
	}

	for _, opt := range opts {
		opt(runOptions)
	}

	z.handler = handler
	z.routes = routes

	zlogger.LogFatal(z.loadConfig(), "ziggurat app load config", nil)

	runOptions.setDefaults()
	z.messageRetry = runOptions.Retry(z.configStore)
	z.httpServer = runOptions.HTTPServer(z.configStore)
	z.metricPublisher = runOptions.MetricPublisher(z.configStore)
	z.httpServer.ConfigureRoutes(z, runOptions.HTTPConfigFunc)
	z.stopFunc = runOptions.StopCallback
	z.startFunc = runOptions.StartCallback

	atomic.StoreInt32(&z.isRunning, 1)
	go func() {
		<-z.start(runOptions.StartCallback)
		z.cancelFun()
		atomic.StoreInt32(&z.isRunning, 0)
		z.stop(z.stopFunc)
		close(z.doneChan)
	}()
	return z.doneChan
}

func (z *Ziggurat) start(startCallback ztype.StartFunction) chan struct{} {

	components := z.components()

	for i, _ := range components {
		c := components[i]
		switch t := c.(type) {
		case ztype.MessageRetry:
			if z.ConfigStore().Config().Retry.Enabled {
				zlogger.LogFatal(c.Start(z), "error starting retries", nil)
			}
		default:
			zlogger.LogError(c.Start(z), "error starting component ", map[string]interface{}{"COMPONENT": t})
		}
	}

	streamsStop, streamsStartErr := kstream.NewKafkaStreams().Start(z)
	if streamsStartErr != nil {
		zlogger.LogFatal(streamsStartErr, "error starting kafka streams", nil)
	}

	startCallback(z)

	return streamsStop

}

func (z *Ziggurat) Stop() {
	z.cancelFun()
}

func (z *Ziggurat) stop(stopFunc ztype.StopFunction) {
	components := z.components()
	for i, _ := range components {
		c := components[i]
		c.Stop(z)
	}
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

func (z *Ziggurat) HTTPServer() ztype.Server {
	return z.httpServer
}

func (z *Ziggurat) ConfigStore() ztype.ConfigStore {
	return z.configStore
}

func (z *Ziggurat) components() []ztype.StartStopper {
	return []ztype.StartStopper{z.metricPublisher, z.messageRetry, z.httpServer}
}

func (z *Ziggurat) IsRunning() bool {
	if atomic.LoadInt32(&z.isRunning) == 1 {
		return true
	}
	return false
}
