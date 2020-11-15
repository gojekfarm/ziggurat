package util

import (
	"github.com/gojekfarm/ziggurat-go/pkg/ziggurat/basic"
	at "github.com/gojekfarm/ziggurat-go/pkg/ziggurat/z"
)

func PipeHandlers(funcs ...at.MiddlewareFunc) func(handlerFunc at.HandlerFunc) at.HandlerFunc {
	return func(next at.HandlerFunc) at.HandlerFunc {
		return func(messageEvent basic.MessageEvent, app at.App) at.ProcessStatus {
			var handlerResult at.HandlerFunc
			last := len(funcs) - 1
			for i := last; i >= 0; i-- {
				f := funcs[i]
				if i == last {
					handlerResult = f(next)
				} else {
					handlerResult = f(handlerResult)
				}
			}
			if handlerResult != nil {
				return handlerResult(messageEvent, app)
			}
			return next(messageEvent, app)
		}
	}
}
