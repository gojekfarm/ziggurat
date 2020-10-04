package ziggurat

type App struct {
	HttpServer      HttpServer
	Retrier         MessageRetrier
	Config          Config
	StreamRouter    *StreamRouter
	MetricPublisher MetricPublisher
}
