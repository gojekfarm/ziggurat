package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojek/ziggurat"
)

func main() {
	sr := ziggurat.NewStreamRouter()
	sr.HandlerFunc("test-entity", func(message *kafka.Message) {
		fmt.Printf("[handlerFunc]: Received message for test-entity1 [key: %s value: %s]\n", message.Key, message.Value)
	})

	sr.HandlerFunc("test-entity2", func(message *kafka.Message) {
		fmt.Printf("[handlerFunc]: Recevied message for test-entity2 [key: %s value: %s]\n", message.Key, message.Value)
	})

	sr.Start(ziggurat.StreamRouterConfigMap{
		"test-entity": ziggurat.StreamRouterConfig{
			InstanceCount:    2,
			BootstrapServers: []string{"localhost:9092"},
			OriginTopics:     []string{"test-topic1"},
			GroupID:          "testGroup",
		},
		"test-entity2": ziggurat.StreamRouterConfig{
			InstanceCount:    2,
			BootstrapServers: []string{"localhost:9092"},
			OriginTopics:     []string{"test-topic2"},
			GroupID:          "testGroup",
		},
	})

}
