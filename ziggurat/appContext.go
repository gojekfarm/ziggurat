package ziggurat

type ApplicationContext struct {
	HttpServer   Http
	Retrier      MessageRetrier
	Config       Config
	StreamRouter *StreamRouter
}
