package util

import (
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
)

var PipeHandlers = func(funcs ...z.MiddlewareFunc) func(handlerFunc z.MessageHandler) z.MessageHandler {
	return func(next z.MessageHandler) z.MessageHandler {
		return z.HandlerFunc(func(messageEvent basic.MessageEvent, app z.App) z.ProcessStatus {
			var handlerResult z.MessageHandler
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
				return handlerResult.HandleMessage(messageEvent, app)
			}
			return next.HandleMessage(messageEvent, app)
		})
	}
}
