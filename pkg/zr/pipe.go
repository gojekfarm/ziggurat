package zr

import (
	"github.com/gojekfarm/ziggurat/pkg/z"
	"github.com/gojekfarm/ziggurat/pkg/zb"
)

var PipeHandlers = func(funcs ...Adapter) func(origHandler z.MessageHandler) z.MessageHandler {
	return func(next z.MessageHandler) z.MessageHandler {
		return z.HandlerFunc(func(messageEvent zb.MessageEvent, app z.App) z.ProcessStatus {
			var handlerResult = next
			lastIdx := len(funcs) - 1
			for i := range funcs {
				f := funcs[lastIdx-i]
				if i == lastIdx-i {
					handlerResult = f(next)
				} else {
					handlerResult = f(handlerResult)
				}
			}
			return handlerResult.HandleMessage(messageEvent, app)
		})
	}
}
