package z

import (
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"net/http"
)

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
