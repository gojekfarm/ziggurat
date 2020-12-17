package ziggurat

var PipeHandlers = func(adapters ...Adapter) func(origHandler MessageHandler) MessageHandler {
	return func(next MessageHandler) MessageHandler {
		return HandlerFunc(func(messageEvent Event, app AppContext) ProcessStatus {
			var handlerResult = next
			lastIdx := len(adapters) - 1
			for i := lastIdx; i >= 0; i-- {
				f := adapters[i]
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
