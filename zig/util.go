package zig

func PipeHandlers(funcs ...MiddlewareFunc) func(handlerFunc HandlerFunc) HandlerFunc {
	return func(origHandler HandlerFunc) HandlerFunc {
		var handlerResult HandlerFunc
		return func(messageEvent MessageEvent, app *App) ProcessStatus {
			last := len(funcs) - 1
			for i := last; i >= 0; i-- {
				f := funcs[i]
				if i == last {
					handlerResult = f(origHandler)
				} else {
					handlerResult = f(handlerResult)
				}
			}
			return handlerResult(messageEvent, app)
		}
	}
}
