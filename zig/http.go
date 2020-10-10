package zig

type HttpServer interface {
	Start(app *App)
	Stop() error
}
