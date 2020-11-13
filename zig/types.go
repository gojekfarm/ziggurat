package zig

import (
	"context"
	"net/http"
)

//public types
type ProcessStatus int
type HandlerFunc func(messageEvent MessageEvent, app App) ProcessStatus
type StartFunction func(a App)
type StopFunction func()
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// Public interfaces
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
	Retry(app App, payload MessageEvent) error
	Stop() error
	Replay(app App, topicEntity string, count int) error
}

type StreamRouter interface {
	Start(app App) (chan int, error)
	HandlerFunc(topicEntityName string, handlerFn HandlerFunc, mw ...MiddlewareFunc)
	GetTopicEntities() []*topicEntity
	GetHandlerFunctionMap() map[string]*topicEntity
	GetTopicEntityNames() []string
}

type App interface {
	Context() context.Context
	Router() StreamRouter
	MessageRetry() MessageRetry
	Run(router StreamRouter, options RunOptions) chan struct{}
	Configure(configFunc func(o App) Options)
	MetricPublisher() MetricPublisher
	HTTPServer() HttpServer
	Config() *Config
	Stop()
	IsRunning() bool
}

type ConfigReader interface {
	Config() *Config
	Parse(options CommandLineOptions)
	GetByKey(key string) interface{}
	Validate(rules map[string]func(c *Config) error) error
	UnmarshalByKey(key string, model interface{}) error
}

// Public constants
const ProcessingSuccess ProcessStatus = 0
const RetryMessage ProcessStatus = 1
const SkipMessage ProcessStatus = 2
