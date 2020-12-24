package ziggurat

import "context"

type Handler interface {
	HandleMessage(event Event) ProcessStatus
}

type Streams interface {
	Consume(ctx context.Context, handler Handler, ) chan error
}

type StructuredLogger interface {
	Info(message string, kvs ...map[string]interface{})
	Debug(message string, kvs ...map[string]interface{})
	Warn(message string, kvs ...map[string]interface{})
	Error(message string, err error, kvs ...map[string]interface{})
	Fatal(message string, err error, kvs ...map[string]interface{})
}

type Event interface {
	Value() []byte
	Key() []byte
	Context() context.Context
	Header(key string) string
}
