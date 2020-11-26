package z

import (
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"net/http"
)

type HandlerFunc func(messageEvent basic.MessageEvent, app App) ProcessStatus

func (h HandlerFunc) HandleMessage(event basic.MessageEvent, app App) ProcessStatus {
	return h(event, app)
}

type StartFunction func(a App)
type StopFunction func()
type MiddlewareFunc func(next MessageHandler) MessageHandler

type Opts = func(ro *RunOptions)

type RunOptions struct {
	HTTPConfigFunc  func(a App, h http.Handler)
	StartCallback   func(a App)
	StopCallback    func()
	HTTPServer      func(c ConfigStore) Server
	Retry           func(c ConfigStore) MessageRetry
	MetricPublisher func(c ConfigStore) MetricPublisher
}

const ProcessingSuccess ProcessStatus = 0
const RetryMessage ProcessStatus = 1
const SkipMessage ProcessStatus = 2

type ProcessStatus int
