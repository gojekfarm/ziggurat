package ziggurat

type MessageRetrier interface {
	Start(app *App) error
	Retry(app *App, payload MessageEvent) error
	Stop() error
	Consume(app *App)
	Replay(app *App, topicEntity string, count int) error
}
