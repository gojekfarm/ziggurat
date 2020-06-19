package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"sync"
)

func main() {
	c1, _ := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})

	c2, _ := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})

	var wg sync.WaitGroup

	_ = c1.SubscribeTopics([]string{"test-go-topic"}, func(consumer *kafka.Consumer, event kafka.Event) error {
		fmt.Println("Rebalance called for c1")
		fmt.Println(consumer.Assignment())
		return nil
	})

	_ = c2.SubscribeTopics([]string{"test-go-topic"}, func(consumer *kafka.Consumer, event kafka.Event) error {
		fmt.Println("Rebalance called for c2")
		fmt.Println(consumer.Assignment())
		return nil
	})


	log.Println("starting kafka consumer")

	wg.Add(2)
	go func() {
		for {
			msg, err := c1.ReadMessage(-1)
			if err == nil {
				fmt.Println("C1 reading a message")
				fmt.Printf("Message on parition  [%d]: %s\n", msg.TopicPartition.Partition, string(msg.Value))
			} else {
				// The client will automatically try to recover from all errors.
				fmt.Println("Error code => ", err.(kafka.Error).Code())
				fmt.Println(kafka.ErrTimedOut)
			}

		}
		wg.Done()
	}()

	go func() {
		for {
			msg, err := c2.ReadMessage(-1)
			if err == nil {
				fmt.Println("C2 reading a message")
				fmt.Printf("Message on partition [%d]: %s\n", msg.TopicPartition.Partition, string(msg.Value))
			} else {
				// The client will automatically try to recover from all errors.
				fmt.Println("Error code => ", err.(kafka.Error).Code())
				fmt.Println(kafka.ErrTimedOut)
			}

		}
		wg.Done()
	}()

	wg.Wait()

}
