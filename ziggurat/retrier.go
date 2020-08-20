package ziggurat

type RetryPayload struct {
	MessageValueBytes []byte
	MessageKeyBytes   []byte
	TopicEntity       string
	RetryCount        int
	MessageTTL        string
}

type MessageRetrier interface {
	Start(config Config, streamRoutes TopicEntityHandlerMap) error
	Retry(config Config, payload MessageEvent) error
	Stop() error
	Consume(config Config, streamRoutes TopicEntityHandlerMap)
}
