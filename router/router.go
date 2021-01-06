package router

import (
	"fmt"
	"github.com/gojekfarm/ziggurat"
)

type defaultRouter struct {
	handlerFunctionMap map[string]ziggurat.HandlerFunc
	l                  ziggurat.StructuredLogger
}

func (dr *defaultRouter) HandleEvent(event ziggurat.Event) ziggurat.ProcessStatus {
	route := event.Headers()[ziggurat.HeaderMessageRoute]
	if handler, ok := dr.handlerFunctionMap[route]; !ok {
		return ziggurat.SkipMessage
	} else {
		return handler.HandleEvent(event)
	}
}

func New() *defaultRouter {
	return &defaultRouter{
		handlerFunctionMap: map[string]ziggurat.HandlerFunc{},
	}
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
