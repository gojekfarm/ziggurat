package router

import (
	"fmt"
	"github.com/gojekfarm/ziggurat"
)

type defaultRouter struct {
	handlerFunctionMap map[string]ziggurat.HandlerFunc
	NotFoundHandler    ziggurat.HandlerFunc
}

func WithNotFoundHandler(nfh func(event ziggurat.Event) ziggurat.ProcessStatus) func(dr *defaultRouter) {
	return func(dr *defaultRouter) {
		dr.NotFoundHandler = nfh
	}
}

func (dr *defaultRouter) HandleEvent(event ziggurat.Event) ziggurat.ProcessStatus {
	route := event.Headers()[ziggurat.HeaderMessageRoute]
	if handler, ok := dr.handlerFunctionMap[route]; !ok {
		return dr.NotFoundHandler.HandleEvent(event)
	} else {
		return handler.HandleEvent(event)
	}
}

func New(opts ...func(dr *defaultRouter)) *defaultRouter {
	dr := &defaultRouter{
		handlerFunctionMap: map[string]ziggurat.HandlerFunc{},
	}
	for _, opt := range opts {
		opt(dr)
	}

	if dr.NotFoundHandler == nil {
		dr.NotFoundHandler = func(event ziggurat.Event) ziggurat.ProcessStatus {
			return ziggurat.SkipMessage
		}
	}

	return dr

}

func (dr *defaultRouter) HandleFunc(route string, handlerFunc func(event ziggurat.Event) ziggurat.ProcessStatus) {
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

func (dr *defaultRouter) Compose(mw ...func(h ziggurat.Handler) ziggurat.Handler) ziggurat.Handler {
	return PipeHandlers(mw...)(dr)
}

func (dr *defaultRouter) Lookup(route string) (ziggurat.Handler, bool) {
	h, ok := dr.handlerFunctionMap[route]
	return h, ok
}
