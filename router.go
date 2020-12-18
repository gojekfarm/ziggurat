package ziggurat

import (
	"context"
)

type defaultRouter struct {
	handlerFunctionMap map[string]HandlerFunc
	l                  LeveledLogger
}

type Adapter func(next MessageHandler) MessageHandler

func (dr *defaultRouter) HandleMessage(event MessageEvent, ctx context.Context) ProcessStatus {
	route := event.StreamRoute
	if handler, ok := dr.handlerFunctionMap[route]; !ok {
		dr.l.Warnf("handler not found, skipping message, ROUTE_NAME=%s", route)
		return SkipMessage
	} else {
		return handler.HandleMessage(event, ctx)
	}
}

func NewRouter() *defaultRouter {
	return &defaultRouter{
		handlerFunctionMap: map[string]HandlerFunc{},
	}
}

func (dr *defaultRouter) HandleFunc(route string, handlerFunc func(event MessageEvent, ctx context.Context) ProcessStatus) {
	if handlerFunc == nil {
		panic("handler cannot be nil")
	}
	dr.handlerFunctionMap[route] = handlerFunc
}

func (dr *defaultRouter) Compose(mw ...Adapter) MessageHandler {
	return PipeHandlers(mw...)(dr)
}

func (dr *defaultRouter) Lookup(route string) (MessageHandler, bool) {
	h, ok := dr.handlerFunctionMap[route]
	return h, ok
}
