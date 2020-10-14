package zig

type MessageRetrier interface {
	Start(app *App) (chan int, error)
	Retry(app *App, payload MessageEvent) error
	Stop() error
	Replay(app *App, topicEntity string, count int) error
}
