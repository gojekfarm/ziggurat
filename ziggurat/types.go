package ziggurat

type HandlerFunc func(message interface{})
type StartFunction func(config ZigguratConfig)
type StopFunction func(config ZigguratConfig)
