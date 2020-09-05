package ziggurat

type Http interface {
	Start(config Config, handlerFuncMap TopicEntityHandlerMap) error
	Stop() error
}
