package ziggurat

type ApplicationContext struct {
	HttpServer   HttpServer
	Retrier      MessageRetrier
	Config       Config
	StreamRouter *StreamRouter
}
