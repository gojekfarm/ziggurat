package z

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"net/http"
)

type TopicEntity struct {
	HandlerFunc      HandlerFunc
	Consumers        []*kafka.Consumer
	bootstrapServers string
	originTopics     []string
	EntityName       string
}

type Middleware = []MiddlewareFunc

type TopicEntityHandlerMap = map[string]*TopicEntity
type HandlerFunc func(messageEvent basic.MessageEvent, app App) ProcessStatus
type StartFunction func(a App)
type StopFunction func()
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

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
	HandlerFunc(topicEntityName string, handlerFn HandlerFunc, mw ...MiddlewareFunc)
	GetTopicEntities() []*TopicEntity
	GetHandlerFunctionMap() map[string]*TopicEntity
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

type RunOptions struct {
	HTTPConfigFunc func(a App, h http.Handler)
	StartCallback  func(a App)
	StopCallback   func()
}

type Options struct {
	HttpServer      HttpServer
	Retrier         MessageRetry
	MetricPublisher MetricPublisher
}

const ProcessingSuccess ProcessStatus = 0
const RetryMessage ProcessStatus = 1
const SkipMessage ProcessStatus = 2

type ProcessStatus int
