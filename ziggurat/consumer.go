package ziggurat

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"sync"
)

type HandlerFunc func(message *kafka.Message)

func StartConsumers(consumerConfig *kafka.ConfigMap, topics []string, instances int, handlerFunc HandlerFunc) {
	var wg sync.WaitGroup
	for i := 0; i < instances; i++ {
		consumer, _ := kafka.NewConsumer(consumerConfig)
		_ = consumer.SubscribeTopics(topics, nil)
		groupID, _ := consumerConfig.Get("group.id", "")
		log.Printf("starting consumer %s-%d\n", groupID, i)
		wg.Add(1)
		go func(c *kafka.Consumer, instanceID int) {
			for {
				msg, err := c.ReadMessage(-1)
				if err != nil && err.(kafka.Error).Code() != kafka.ErrTimedOut {
					break
				}
				if msg != nil {
					log.Printf("Received message by %s-%d on partition %d\n", groupID, instanceID, msg.TopicPartition.Partition)
					handlerFunc(msg)
					_, commitErr := c.CommitMessage(msg)
					if commitErr != nil {
						log.Printf("error committing message %v\n", commitErr)
					}
					log.Printf("committed message successfully\n")
				}
			}
			wg.Done()
		}(consumer, i)
	}
	wg.Wait()
}
