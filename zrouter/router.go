package zrouter

import (
	"errors"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/zlog"
	"github.com/gojekfarm/ziggurat/ztype"
)

type defaultRouter struct {
	handlerFunctionMap map[string]ztype.HandlerFunc
}

type Adapter func(next ztype.MessageHandler) ztype.MessageHandler

func (dr *defaultRouter) HandleMessage(event zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
	route := event.StreamRoute
	if handler, ok := dr.handlerFunctionMap[route]; !ok {
		zlog.LogWarn("handler not found, skipping message", map[string]interface{}{"ROUTE": route})
		return ztype.SkipMessage
	} else {
		return handler.HandleMessage(event, app)
	}
}

func New() *defaultRouter {
	return &defaultRouter{
		handlerFunctionMap: map[string]ztype.HandlerFunc{},
	}
}

func (dr *defaultRouter) HandleFunc(route string, handlerFunc func(event zbase.MessageEvent, app ztype.App) ztype.ProcessStatus) {
	if handlerFunc == nil {
		zlog.LogFatal(errors.New("handler cannot be nil"), "router error", map[string]interface{}{"ROUTE": route})
	}
	dr.handlerFunctionMap[route] = handlerFunc
}

func (dr *defaultRouter) Compose(mw ...Adapter) ztype.MessageHandler {
	return PipeHandlers(mw...)(dr)
}

func (dr *defaultRouter) Routes() []string {
	routes := []string{}
	for route, _ := range dr.handlerFunctionMap {
		routes = append(routes, route)
	}
	return routes
}

func (dr *defaultRouter) Lookup(route string) (ztype.MessageHandler, bool) {
	h, ok := dr.handlerFunctionMap[route]
	return h, ok
}
