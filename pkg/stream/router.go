package stream

import (
	"github.com/gojekfarm/ziggurat-go/pkg/util"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
)

type defaultRouter struct {
	handlerFunctionMap map[string]z.HandlerFunc
	routeMiddlewareMap map[string][]z.MiddlewareFunc
	routes             []string
}

func NewRouter() *defaultRouter {
	return &defaultRouter{
		handlerFunctionMap: map[string]z.HandlerFunc{},
		routes:             []string{},
		routeMiddlewareMap: map[string][]z.MiddlewareFunc{},
	}
}

func (dr *defaultRouter) RouteHandlerMap() map[string]z.HandlerFunc {
	return dr.handlerFunctionMap
}

func (dr *defaultRouter) Routes() []string {
	return dr.routes
}

func (dr *defaultRouter) Build() {
	for route, middlewareFuncs := range dr.routeMiddlewareMap {
		hf := dr.handlerFunctionMap[route]
		dr.handlerFunctionMap[route] = util.PipeHandlers(middlewareFuncs...)(hf)
	}
}

func (dr *defaultRouter) HandlerFunc(route string, handlerFn z.HandlerFunc, mw ...z.MiddlewareFunc) *defaultRouter {
	dr.routes = append(dr.routes, route)
	if len(mw) > 0 {
		dr.routeMiddlewareMap[route] = mw
	}
	dr.handlerFunctionMap[route] = handlerFn
	return dr
}

func (dr *defaultRouter) Use(middlewareFunc ...z.MiddlewareFunc) *defaultRouter {
	for route, handler := range dr.handlerFunctionMap {
		origHandler := handler
		if len(middlewareFunc) > 0 {
			dr.handlerFunctionMap[route] = util.PipeHandlers(middlewareFunc...)(origHandler)
		}
	}
	return dr
}
