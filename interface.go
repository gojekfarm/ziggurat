package ziggurat

import "context"

type MessageHandler interface {
	HandleMessage(event MessageEvent, ctx context.Context) ProcessStatus
}

type Streams interface {
	Consume(ctx context.Context, routes Routes, handler MessageHandler) chan error
}
