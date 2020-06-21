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

func startConsumer(topicEntity TopicEntityName, handlerFunc HandlerFunc, wg *sync.WaitGroup, consumer *kafka.Consumer, instanceID string) {
	go func(c *kafka.Consumer, instanceID string, waitGroup *sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()
		log.Printf("starting consumer %s %s\n", topicEntity, instanceID)
		for {
			msg, err := c.ReadMessage(-1)
			if err != nil && err.(kafka.Error).Code() != kafka.ErrTimedOut {
				fmt.Printf("[%s] error: %v\n ", instanceID, err.(kafka.Error).Code())
				break
			}
			if msg != nil {
				log.Printf("Received message by %s on partition %s-%d\n", instanceID, *msg.TopicPartition.Topic, msg.TopicPartition.Partition)
				handlerFunc(msg)
				_, commitErr := c.CommitMessage(msg)
				if commitErr != nil {
					log.Printf("error committing message %v\n", commitErr)
				}
				log.Printf("committed message successfully\n")
			}
		}
		consumer.Close()
	}(consumer, instanceID, wg)
}

func StartConsumers(consumerConfig *kafka.ConfigMap, topicEntity TopicEntityName, topics []string, instances int, handlerFunc HandlerFunc, wg *sync.WaitGroup) {
	for i := 0; i < instances; i++ {
		consumer := createConsumer(consumerConfig, topics)
		groupID, _ := consumerConfig.Get("group.id", "")
		instanceID := fmt.Sprintf("%s-%s-%d", topicEntity, groupID, i)
		startConsumer(topicEntity, handlerFunc, wg, consumer, instanceID)
	}
}
