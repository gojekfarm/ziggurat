package ziggurat

import (
	"context"
	"time"
)

type Routes map[string]RouteDefinition

type AppContext interface {
	Context() context.Context
	Routes() Routes
	Handler() MessageHandler
}

type MessageHandler interface {
	HandleMessage(event Event, app AppContext) ProcessStatus
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

type Event interface {
	PublishTime() time.Time
	MessageKey() []byte
	MessageValue() []byte
	OriginTopic() string
	RouteName() string
	GetAttribute(key string) interface{}
	SetAttribute(key string, value interface{})
	Version() string
}
