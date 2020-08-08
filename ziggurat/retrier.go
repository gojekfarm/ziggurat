package ziggurat

type RetryPayload struct {
	MessageValueBytes []byte
	MessageKeyBytes   []byte
	RetryAttributes   map[string]interface{}
}

type MessageRetrier interface {
	Start(config Config) error
	Retry(payload RetryPayload) error
	Consume(handlerFunc HandlerFunc)
	Stop() error
}

