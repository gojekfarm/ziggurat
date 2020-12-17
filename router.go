package ziggurat

import (
	"errors"
)

type defaultRouter struct {
	handlerFunctionMap map[string]HandlerFunc
}

type Adapter func(next MessageHandler) MessageHandler

func (dr *defaultRouter) HandleMessage(event MessageEvent, app AppContext) ProcessStatus {
	route := event.StreamRoute
	if handler, ok := dr.handlerFunctionMap[route]; !ok {
		LogWarn("handler not found, skipping message", map[string]interface{}{"ROUTE": route})
		return SkipMessage
	} else {
		return handler.HandleMessage(event, app)
	}
}

func NewRouter() *defaultRouter {
	return &defaultRouter{
		handlerFunctionMap: map[string]HandlerFunc{},
	}
}

func (dr *defaultRouter) HandleFunc(route string, handlerFunc func(event MessageEvent, app AppContext) ProcessStatus) {
	if handlerFunc == nil {
		LogFatal(errors.New("handler cannot be nil"), "router error", map[string]interface{}{"ROUTE": route})
	}
	dr.handlerFunctionMap[route] = handlerFunc
}

func (dr *defaultRouter) Compose(mw ...Adapter) MessageHandler {
	return PipeHandlers(mw...)(dr)
}

func (dr *defaultRouter) Routes() []string {
	routes := []string{}
	for route, _ := range dr.handlerFunctionMap {
		routes = append(routes, route)
	}
	return routes
}

func (dr *defaultRouter) Lookup(route string) (MessageHandler, bool) {
	h, ok := dr.handlerFunctionMap[route]
	return h, ok
}
