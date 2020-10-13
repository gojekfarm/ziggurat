package zig

type HandlerFunc func(messageEvent MessageEvent, app *App) ProcessStatus
type StartFunction func(a *App)
type StopFunction func()
type ProcessStatus int
type Middleware func(next HandlerFunc) HandlerFunc

const ProcessingSuccess ProcessStatus = 0
const RetryMessage ProcessStatus = 1
const SkipMessage ProcessStatus = 2
