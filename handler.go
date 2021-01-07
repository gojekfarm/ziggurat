package ziggurat

import "context"

type ProcessStatus int

const ProcessingSuccess ProcessStatus = 0
const RetryMessage ProcessStatus = 1
const SkipMessage ProcessStatus = 2

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
