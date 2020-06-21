package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"strings"
	"sync"
)

type HandlerFunc func(message *kafka.Message)

func startConsumers(consumerConfig *kafka.ConfigMap, topics []string, instances int, handlerFunc HandlerFunc) {
	var wg sync.WaitGroup
	for i := 0; i < instances; i++ {
		consumer, _ := kafka.NewConsumer(consumerConfig)
		_ = consumer.SubscribeTopics(topics, nil)
		log.Printf("%v subscribed to %s topics\n", consumerConfig.Get("group.id", ""), strings.Join(topics, ","))
		wg.Add(1)
		go func() {
			for {
				msg, err := consumer.ReadMessage(-1)
				if err != nil && err.(kafka.Error).Code() != kafka.ErrTimedOut {
					break
				}
				if msg != nil {
					log.Printf("Received message at %s", msg.Timestamp)
					handlerFunc(msg)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func main() {
	consumerConfig := &kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	}
	startConsumers(consumerConfig, []string{"test-topic"}, 1, func(message *kafka.Message) {
		fmt.Printf("Recevied message %s", message.Value)
	})

}
