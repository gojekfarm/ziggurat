package ziggurat

var PipeHandlers = func(funcs ...func(handler Handler) Handler) func(origHandler Handler) Handler {
	return func(next Handler) Handler {
		return HandlerFunc(func(event Event) ProcessStatus {
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
			return handlerResult.HandleEvent(event)
		})
	}
}
