package z

import (
	"context"
	"github.com/gojekfarm/ziggurat-go/pkg/zbasic"
	"net/http"
)

type StartStopper interface {
	Start(a App) error
	Stop(a App)
}

type Server interface {
	ConfigureRoutes(a App, configFunc func(a App, h http.Handler))
	Handler() http.Handler
	StartStopper
}

type MetricPublisher interface {
	StartStopper
	IncCounter(metricName string, value int64, arguments map[string]string) error
	Gauge(metricName string, value int64, arguments map[string]string) error
}

type MessageRetry interface {
	Retry(app App, payload zbasic.MessageEvent) error
	Replay(app App, topicEntity string, count int) error
	StartStopper
}

type ConfigStore interface {
	Config() *zbasic.Config
	Parse(options zbasic.CommandLineOptions) error
	GetByKey(key string) interface{}
	UnmarshalByKey(key string, model interface{}) error
}

type App interface {
	Context() context.Context
	Routes() []string
	MessageRetry() MessageRetry
	Handler() MessageHandler
	MetricPublisher() MetricPublisher
	HTTPServer() Server
	ConfigStore() ConfigStore
}

type MessageHandler interface {
	HandleMessage(event zbasic.MessageEvent, app App) ProcessStatus
}

type ConfigValidator interface {
	Validate(config *zbasic.Config) error
}
