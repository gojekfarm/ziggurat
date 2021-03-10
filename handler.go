package ziggurat

import "context"

type HandlerFunc func(ctx context.Context, event Event) error

func (h HandlerFunc) Handle(ctx context.Context, event Event) error {
	return h(ctx, event)
}

type Handler interface {
	Handle(ctx context.Context, event Event) error
}
