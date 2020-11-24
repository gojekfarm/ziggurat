package z

import (
	"context"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"net/http"
)

type HttpServer interface {
	Start(app App)
	ConfigureHTTPRoutes(a App, configFunc func(a App, h http.Handler))
	Stop(app App) error
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

type ConfigReader interface {
	Config() *basic.Config
	Parse(options basic.CommandLineOptions)
	GetByKey(key string) interface{}
	Validate(rules map[string]func(c *basic.Config) error) error
	UnmarshalByKey(key string, model interface{}) error
}

type App interface {
	Context() context.Context
	Routes() []string
	MessageRetry() MessageRetry
	Handler() MessageHandler
	Run(router MessageHandler, routes []string, options RunOptions) chan struct{}
	MetricPublisher() MetricPublisher
	HTTPServer() HttpServer
	Config() *basic.Config
	ConfigReader() ConfigReader
}

type MessageHandler interface {
	HandleMessage(event basic.MessageEvent, app App) ProcessStatus
}
