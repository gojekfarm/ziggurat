package ziggurat

type HttpServer interface {
	Start(app *App)
	Stop() error
}
