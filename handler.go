package ziggurat

import "context"

type HandlerFunc func(ctx context.Context, event Event) error

func (h HandlerFunc) HandleEvent(ctx context.Context, event Event) error {
	return h(ctx, event)
}

type Handler interface {
	HandleEvent(ctx context.Context, event Event) error
}
