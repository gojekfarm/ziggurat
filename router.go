package ziggurat

import (
	"context"
)

type defaultRouter struct {
	handlerFunctionMap map[string]HandlerFunc
	l                  StructuredLogger
}

type Adapter func(next Handler) Handler

func (dr *defaultRouter) HandleMessage(event Message, ctx context.Context) ProcessStatus {
	route := event.RoutingKey
	if handler, ok := dr.handlerFunctionMap[route]; !ok {
		dr.l.Warn("handler not found", map[string]interface{}{"ROUTE": route})
		return SkipMessage
	} else {
		return handler.HandleMessage(event, ctx)
	}
}

func NewRouter() *defaultRouter {
	return &defaultRouter{
		handlerFunctionMap: map[string]HandlerFunc{},
		l:                  NewLogger("info"),
	}
}

func (dr *defaultRouter) HandleFunc(route string, handlerFunc func(event Message, ctx context.Context) ProcessStatus) {
	if handlerFunc == nil {
		panic("handler cannot be nil")
	}
	dr.handlerFunctionMap[route] = handlerFunc
}

func (dr *defaultRouter) Compose(mw ...Adapter) Handler {
	return PipeHandlers(mw...)(dr)
}

func (dr *defaultRouter) Lookup(route string) (Handler, bool) {
	h, ok := dr.handlerFunctionMap[route]
	return h, ok
}
