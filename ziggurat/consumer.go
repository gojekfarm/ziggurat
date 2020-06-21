package ziggurat

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"sync"
)

func createConsumer(consumerConfig *kafka.ConfigMap, topics []string) *kafka.Consumer {
	consumer, _ := kafka.NewConsumer(consumerConfig)
	_ = consumer.SubscribeTopics(topics, nil)
	return consumer
}

func StartConsumers(consumerConfig *kafka.ConfigMap, topicEntity TopicEntityName, topics []string, instances int, handlerFunc HandlerFunc, wg *sync.WaitGroup) {
	for i := 0; i < instances; i++ {
		consumer := createConsumer(consumerConfig, topics)
		groupID, _ := consumerConfig.Get("group.id", "")
		wg.Add(1)
		go func(c *kafka.Consumer, instanceID int, waitGroup *sync.WaitGroup) {
			log.Printf("starting consumer %s-%s-%d\n", topicEntity, groupID, instanceID)
			for {
				msg, err := c.ReadMessage(-1)
				if err != nil && err.(kafka.Error).Code() != kafka.ErrTimedOut {
					fmt.Printf("error: %v ", err.(kafka.Error).Code())
					break
				}
				if msg != nil {
					log.Printf("Received message by %s-%d on partition %v-%d\n", groupID, instanceID, *msg.TopicPartition.Topic, msg.TopicPartition.Partition)
					handlerFunc(msg)
					_, commitErr := c.CommitMessage(msg)
					if commitErr != nil {
						log.Printf("error committing message %v\n", commitErr)
					}
					log.Printf("committed message successfully\n")
				}
			}
			wg.Done()
		}(consumer, i, wg)
	}
}
