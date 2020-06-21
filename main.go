package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojek/ziggurat"
)

func main() {
	consumerConfig := &kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	}
	ziggurat.StartConsumers(consumerConfig, []string{"test-topic"}, 3, func(message *kafka.Message) {
		fmt.Printf("Recevied message [ %s %s ]\n", message.Key, message.Value)
	})

}
