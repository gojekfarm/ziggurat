package z

import (
	"context"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"net/http"
)

type HttpServer interface {
	Start(app App)
	ConfigureHTTPRoutes(a App, configFunc func(a App, h http.Handler))
	Stop(ctx context.Context) error
}

type MetricPublisher interface {
	Start(app App) error
	Stop() error
	IncCounter(metricName string, value int64, arguments map[string]string) error
	Gauge(metricName string, value int64, arguments map[string]string) error
}

type MessageRetry interface {
	Start(app App) error
	Retry(app App, payload basic.MessageEvent) error
	Stop() error
	Replay(app App, topicEntity string, count int) error
}

type StreamRouter interface {
	Start(app App) (chan struct{}, error)
	HandlerFunc(topicEntityName string, handlerFn HandlerFunc)
	GetTopicEntities() []*TopicEntity
	GetHandlerFunctionMap() map[string]*TopicEntity
	Use(middlewareFunc MiddlewareFunc, excludeFunc ExcludeFunc)
	GetTopicEntityNames() []string
}

type ConfigReader interface {
	Config() *basic.Config
	Parse(options basic.CommandLineOptions)
	GetByKey(key string) interface{}
	Validate(rules map[string]func(c *basic.Config) error) error
	UnmarshalByKey(key string, model interface{}) error
}

type App interface {
	Context() context.Context
	Router() StreamRouter
	MessageRetry() MessageRetry
	Run(router StreamRouter, options RunOptions) chan struct{}
	Configure(configFunc func(o App) Options)
	MetricPublisher() MetricPublisher
	HTTPServer() HttpServer
	Config() *basic.Config
	ConfigReader() ConfigReader
	Stop()
	IsRunning() bool
}

type StreamWorker interface {
	Push(app App, event basic.MessageEvent) error
	Start(app App)
	Stop()
}
