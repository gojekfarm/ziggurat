package ziggurat

import "github.com/confluentinc/confluent-kafka-go/kafka"

type TopicEntityHandlerMap = map[string]TopicEntity

type TopicEntity struct {
	topicEntity string
	groupId     string
	handlerFunc func(message kafka.Message)
}

type StreamRouter struct {
	handlerFunctionMap map[string]TopicEntity
}

func NewStreamRouter() *StreamRouter {
	return &StreamRouter{
		handlerFunctionMap: make(TopicEntityHandlerMap),
	}
}

func (sr *StreamRouter) HandlerFunc(topicEntity string, groupId string, handlerFn func(message kafka.Message)) {
	sr.handlerFunctionMap[topicEntity] = TopicEntity{
		topicEntity: topicEntity,
		groupId:     groupId,
		handlerFunc: handlerFn,
	}
}
