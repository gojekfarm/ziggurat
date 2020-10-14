package zig

//public types
type ProcessStatus int
type HandlerFunc func(messageEvent MessageEvent, app *App) ProcessStatus
type StartFunction func(a *App)
type StopFunction func()
type Middleware func(next HandlerFunc) HandlerFunc

// Public interfaces
type HttpServer interface {
	Start(app *App)
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

// Public constants
const ProcessingSuccess ProcessStatus = 0
const RetryMessage ProcessStatus = 1
const SkipMessage ProcessStatus = 2
