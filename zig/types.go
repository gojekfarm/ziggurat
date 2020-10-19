package zig

import "github.com/julienschmidt/httprouter"

//public types
type ProcessStatus int
type HandlerFunc func(messageEvent MessageEvent, app *App) ProcessStatus
type StartFunction func(a *App)
type StopFunction func()
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// Public interfaces
type HttpServer interface {
	Start(app *App)
	attachRoute(attachFunc func(r *httprouter.Router))
	Stop() error
}

type MetricPublisher interface {
	Start(app *App) error
	Stop() error
	IncCounter(metricName string, value int, arguments map[string]string) error
}

type MessageRetrier interface {
	Start(app *App) (chan int, error)
	Retry(app *App, payload MessageEvent) error
	Stop() error
	Replay(app *App, topicEntity string, count int) error
}

type StreamRouter interface {
	Start(app *App) (chan int, error)
	HandlerFunc(topicEntityName string, handlerFn HandlerFunc, mw ...Middleware)
	GetTopicEntities() []*topicEntity
	GetHandlerFunctionMap() map[string]*topicEntity
}

// Public constants
const ProcessingSuccess ProcessStatus = 0
const RetryMessage ProcessStatus = 1
const SkipMessage ProcessStatus = 2
