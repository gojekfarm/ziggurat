package ziggurat

type ApplicationContext struct {
	httpServer   *Http
	retrier      *MessageRetrier
	config       *Config
	streamRoutes TopicEntityHandlerMap
}

func (ac *ApplicationContext) attachHttpServer(server *Http) {
	ac.httpServer = server
}

func (ac *ApplicationContext) attachRetrier(retrier *MessageRetrier) {
	ac.retrier = retrier
}

func (ac *ApplicationContext) getHttpServer() *Http {
	return ac.httpServer
}

func (ac *ApplicationContext) getRetrier() *MessageRetrier {
	return ac.retrier
}
