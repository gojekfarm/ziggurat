package ziggurat

type HandlerFunc func(message interface{})

type TopicEntityName string

type StreamRouterConfigMap map[TopicEntityName]StreamRouterConfig
