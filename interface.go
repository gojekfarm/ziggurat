package ziggurat

import (
	"context"
)

type Routes map[string]RouteDefinition

type AppContext interface {
	Context() context.Context
	Routes() Routes
	Handler() MessageHandler
}

type MessageHandler interface {
	HandleMessage(event MessageEvent, app AppContext) ProcessStatus
}

type Streams interface {
	Start(a AppContext) (chan struct{}, error)
	Stop()
}

type RouteDefinition interface {
	ThreadCount() int
	Servers() string
	ConsumerGroupID() string
	Topics() string
}
