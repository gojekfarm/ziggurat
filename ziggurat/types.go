package ziggurat

type HandlerFunc func(messageEvent MessageEvent)
type StartFunction func(config Config)
type StopFunction func()
