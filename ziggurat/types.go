package ziggurat

import "github.com/confluentinc/confluent-kafka-go/kafka"

type HandlerFunc func(message *kafka.Message)

type TopicEntityName string

type StreamRouterConfigMap map[TopicEntityName]StreamRouterConfig
