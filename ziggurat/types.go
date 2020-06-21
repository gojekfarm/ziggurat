package ziggurat

import "github.com/confluentinc/confluent-kafka-go/kafka"

type HandlerFunc func(message *kafka.Message)

type TopicEntityName string

type StreamRouterConfig struct {
	InstanceCount    int
	BootstrapServers []string
	OriginTopics     []string
	GroupID          string
}

type StreamRouterConfigMap map[TopicEntityName]StreamRouterConfig
