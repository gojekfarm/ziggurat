package ziggurat

import "context"

type Handler interface {
	HandleMessage(event *Message, ctx context.Context) ProcessStatus
}

type Streams interface {
	Consume(ctx context.Context, routes Routes, handler Handler) chan error
}

type StructuredLogger interface {
	Info(message string, kvs ...map[string]interface{})
	Debug(message string, kvs ...map[string]interface{})
	Warn(message string, kvs ...map[string]interface{})
	Error(message string, err error, kvs ...map[string]interface{})
	Fatal(message string, err error, kvs ...map[string]interface{})
}
