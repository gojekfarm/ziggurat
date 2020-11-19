package z

import (
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

type RunOptions struct {
	HTTPConfigFunc func(a App, h http.Handler)
	StartCallback  func(a App)
	StopCallback   func()
}

type Options struct {
	HttpServer      HttpServer
	Retry           MessageRetry
	MetricPublisher MetricPublisher
}

const ProcessingSuccess ProcessStatus = 0
const RetryMessage ProcessStatus = 1
const SkipMessage ProcessStatus = 2

type ProcessStatus int
