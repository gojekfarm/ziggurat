package ziggurat

import (
	"context"
)

type Middleware func(handler Handler) Handler

var pipe = func(h Handler, fs ...Middleware) Handler {
	if len(fs) < 1 {
		return h
	}
	last := len(fs) - 1
	f := func(ctx context.Context, event *Event) error {
		next := h
		for i := last; i >= 0; i-- {
			next = fs[i](next)
		}
		return next.Handle(ctx, event)
	}
	return HandlerFunc(f)
}

// Use takes a ziggurat.Handler and wraps it with Middleware
func Use(h Handler, fs ...Middleware) Handler {
	return pipe(h, fs...)
}
