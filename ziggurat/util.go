package ziggurat

func pipeHandlers(funcs ...MiddlewareFunc) func(handlerFunc HandlerFunc) HandlerFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(messageEvent MessageEvent, app App) ProcessStatus {
			var handlerResult HandlerFunc
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
