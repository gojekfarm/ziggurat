package zig

func pipeHandlers(funcs ...Middleware) func(handlerFunc HandlerFunc) HandlerFunc {
	return func(origHandler HandlerFunc) HandlerFunc {
		var handlerResult HandlerFunc
		return func(messageEvent MessageEvent, app *App) ProcessStatus {
			for index, f := range funcs {
				if index == 0 {
					handlerResult = f(origHandler)
				} else {
					handlerResult = f(handlerResult)
				}
			}
			return handlerResult(messageEvent, app)
		}
	}
}
