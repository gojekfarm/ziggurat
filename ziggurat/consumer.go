package ziggurat

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"sync"
	"time"
)

const defaultPollTimeout = 10 * time.Second

func createConsumer(consumerConfig *kafka.ConfigMap, topics []string) *kafka.Consumer {
	consumer, _ := kafka.NewConsumer(consumerConfig)
	_ = consumer.SubscribeTopics(topics, nil)
	return consumer
}

func startConsumer(handlerFunc HandlerFunc, consumer *kafka.Consumer, instanceID string, wg *sync.WaitGroup) {
	log.Printf("starting consumer instance: %s", instanceID)
	go func(c *kafka.Consumer, instanceID string, waitGroup *sync.WaitGroup) {
		for {
			msg, err := c.ReadMessage(defaultPollTimeout)
			if err != nil && err.(kafka.Error).Code() != kafka.ErrTimedOut {
				log.Printf("err on instance %s -> %v", instanceID, err)
			} else if err != nil && err.(kafka.Error).Code() == kafka.ErrAllBrokersDown {
				break
			}

			if msg != nil {
				log.Printf("Message received by [CONSUMER: %s TOPIC: %s PARTITION: %d]\n", instanceID, *msg.TopicPartition.Topic, msg.TopicPartition.Partition)
				handlerFunc(msg)
				fmt.Println(msg.TopicPartition.Offset)
				_, cmtErr := consumer.CommitMessage(msg)
				if cmtErr != nil {
					break
				}
			}
		}
		wg.Done()
	}(consumer, instanceID, wg)
}

func StartConsumers(consumerConfig *kafka.ConfigMap, topicEntity TopicEntityName, topics []string, instances int, handlerFunc HandlerFunc, wg *sync.WaitGroup) []*kafka.Consumer {
	consumers := make([]*kafka.Consumer, 0, instances)
	for i := 0; i < instances; i++ {
		consumer := createConsumer(consumerConfig, topics)
		consumers = append(consumers, consumer)
		groupID, _ := consumerConfig.Get("group.id", "")
		instanceID := fmt.Sprintf("%s-%s-%d", topicEntity, groupID, i)
		wg.Add(1)
		startConsumer(handlerFunc, consumer, instanceID, wg)
	}
	return consumers
}
