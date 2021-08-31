package router

import (
	"context"
	"fmt"

	"github.com/gojekfarm/ziggurat"
)

type RouteNotFoundError struct {
	routeName string
}

func (r RouteNotFoundError) Error() string {
	return fmt.Sprintf("route %s not found", r.routeName)
}

type defaultRouter struct {
	handlerFunctionMap map[string]ziggurat.HandlerFunc
	NotFoundHandler    ziggurat.HandlerFunc
}

func WithNotFoundHandler(nfh func(ctx context.Context, event *ziggurat.Event) error) func(dr *defaultRouter) {
	return func(dr *defaultRouter) {
		dr.NotFoundHandler = nfh
	}
}

func (dr *defaultRouter) Handle(ctx context.Context, event *ziggurat.Event) error {
	route := event.Path
	if handler, ok := dr.handlerFunctionMap[route]; !ok {
		return dr.NotFoundHandler.Handle(ctx, event)
	} else {
		return handler.Handle(ctx, event)
	}
}

// New created a new router
// router stores handlers in a map[string]ziggurat.Handler
// the route received in the event header is matched against the entries in the map
// and the corresponding handler is executed
func New(opts ...func(dr *defaultRouter)) *defaultRouter {
	dr := &defaultRouter{
		handlerFunctionMap: map[string]ziggurat.HandlerFunc{},
	}
	for _, opt := range opts {
		opt(dr)
	}

	if dr.NotFoundHandler == nil {
		dr.NotFoundHandler = func(ctx context.Context, event *ziggurat.Event) error {
			return RouteNotFoundError{routeName: event.Path}
		}
	}
	return dr

}

// HandleFunc registers a new handler function for the given route
func (dr *defaultRouter) HandleFunc(route string, handlerFunc func(ctx context.Context, event *ziggurat.Event) error) {
	if route == "" {
		panic(`route cannot be ""`)
	}
	if handlerFunc == nil {
		panic("handler cannot be nil")
	}

	if _, ok := dr.handlerFunctionMap[route]; ok {
		panic(fmt.Sprintf("route %s has already been registered", route))
	}
	dr.handlerFunctionMap[route] = handlerFunc
}

// Lookup looks up and returns a handlerFunc if found
// it returns a second bool value `true` if found `false` if not found
func (dr *defaultRouter) Lookup(route string) (ziggurat.Handler, bool) {
	h, ok := dr.handlerFunctionMap[route]
	return h, ok
}
