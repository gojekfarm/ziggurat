package zrouter

import (
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/ztype"
)

var PipeHandlers = func(funcs ...Adapter) func(origHandler ztype.MessageHandler) ztype.MessageHandler {
	return func(next ztype.MessageHandler) ztype.MessageHandler {
		return ztype.HandlerFunc(func(messageEvent zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
			var handlerResult = next
			lastIdx := len(funcs) - 1
			for i := lastIdx; i >= 0; i-- {
				f := funcs[i]
				if i == lastIdx {
					handlerResult = f(next)
				} else {
					handlerResult = f(handlerResult)
				}
			}
			return handlerResult.HandleMessage(messageEvent, app)
		})
	}
}
