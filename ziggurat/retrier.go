package ziggurat

type RetryPayload struct {
	MessageValueBytes []byte
	MessageKeyBytes   []byte
	RetryAttributes   map[string]interface{}
}

type MessageRetrier interface {
	Start(config Config, streamRoutes TopicEntityHandlerMap) error
	Retry(payload RetryPayload) error
	Stop() error
	Consume(config Config, streamRoutes TopicEntityHandlerMap)
}
