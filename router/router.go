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

func WithNotFoundHandler(nfh func(ctx context.Context, event ziggurat.Event) error) func(dr *defaultRouter) {
	return func(dr *defaultRouter) {
		dr.NotFoundHandler = nfh
	}
}

func (dr *defaultRouter) Handle(ctx context.Context, event ziggurat.Event) error {
	route := event.Headers()[ziggurat.HeaderMessageRoute]
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
		dr.NotFoundHandler = func(ctx context.Context, event ziggurat.Event) error {
			return RouteNotFoundError{routeName: event.Headers()[ziggurat.HeaderMessageRoute]}
		}
	}
	return dr

}

// HandleFunc registers a new handler function for the given route
func (dr *defaultRouter) HandleFunc(route string, handlerFunc func(ctx context.Context, event ziggurat.Event) error) {
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

// Compose takes in a set of middleware functions
// builds a chain of executing from left to right and returns a new handler
// router := New()
// router.HandleFunc("my-route",func(c context.Context,e ziggurat.Event) error{
//		return nil
//})
// h := router.Compose(middlewareOne,middlewareTwo,middlewareThree)
// order of execution: middlewareOne -> middlewareTwo -> middlewareThree -> handlerFunc
func (dr *defaultRouter) Compose(mw ...func(h ziggurat.Handler) ziggurat.Handler) ziggurat.Handler {
	return PipeHandlers(mw...)(dr)
}

// Lookup looks up and returns a handlerFunc if found
// it returns a second bool value `true` if found `false` if not found
func (dr *defaultRouter) Lookup(route string) (ziggurat.Handler, bool) {
	h, ok := dr.handlerFunctionMap[route]
	return h, ok
}
