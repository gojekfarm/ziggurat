package ziggurat

import "context"

type HandlerFunc func(event Event) ProcessStatus

func (h HandlerFunc) HandleEvent(event Event) ProcessStatus {
	return h(event)
}

type Handler interface {
	HandleEvent(event Event) ProcessStatus
}

type Event interface {
	Value() []byte
	Headers() map[string]string
	Context() context.Context
}