package zr

import (
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zbasic"
	"github.com/gojekfarm/ziggurat-go/pkg/zlogger"
)

type defaultRouter struct {
	handlerFunctionMap map[string]z.HandlerFunc
}

type Adapter func(next z.MessageHandler) z.MessageHandler

func (dr *defaultRouter) HandleMessage(event zbasic.MessageEvent, app z.App) z.ProcessStatus {
	route := event.StreamRoute
	if handler, ok := dr.handlerFunctionMap[route]; !ok {
		zlogger.LogWarn("handler not found, skipping message", map[string]interface{}{"ROUTE": route})
		return z.SkipMessage
	} else {
		return handler.HandleMessage(event, app)
	}
}

func NewRouter() *defaultRouter {
	return &defaultRouter{
		handlerFunctionMap: map[string]z.HandlerFunc{},
	}
}

func (dr *defaultRouter) HandleFunc(route string, handlerFunc func(event zbasic.MessageEvent, app z.App) z.ProcessStatus) {
	dr.handlerFunctionMap[route] = handlerFunc
}

func (dr *defaultRouter) Compose(mw ...Adapter) z.MessageHandler {
	return PipeHandlers(mw...)(dr)
}

func (dr *defaultRouter) Routes() []string {
	routes := []string{}
	for route, _ := range dr.handlerFunctionMap {
		routes = append(routes, route)
	}
	return routes
}

func (dr *defaultRouter) Lookup(route string) z.MessageHandler {
	return dr.handlerFunctionMap[route]
}
