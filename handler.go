package ziggurat

import "context"

type HandlerFunc func(event Event, ctx context.Context) error

func (h HandlerFunc) HandleEvent(event Event, ctx context.Context) error {
	return h(event, ctx)
}

type Handler interface {
	HandleEvent(event Event, ctx context.Context) error
}
