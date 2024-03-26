package ziggurat

import "context"

// HandlerFunc serves as an adapter to convert
// regular functions of the signature f(context.Context,ziggurat.Event)
// to implement the ziggurat.Handler interface
type HandlerFunc func(ctx context.Context, event *Event)

func (h HandlerFunc) Handle(ctx context.Context, event *Event) {
	h(ctx, event)
}

// Handler is an interface which can be implemented
// to handle messages produced by a stream source
// the router package implements this interface
type Handler interface {
	Handle(ctx context.Context, event *Event)
}
