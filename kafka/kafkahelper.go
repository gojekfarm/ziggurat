package kafka

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func createConsumer(consumerConfig *kafka.ConfigMap, topics []string) confluentConsumer {
	consumer, err := kafka.NewConsumer(consumerConfig)
	if err != nil {
		panic("error creating consumer:" + err.Error())
	}
	subscribeErr := consumer.SubscribeTopics(topics, nil)
	if subscribeErr != nil {
		panic("error subscribing to topics:" + subscribeErr.Error())
	}
	return consumer
}

// storeOffsets ensures at least once delivery
// offsets are stored in memory and are later flushed by the auto-commit timer
func storeOffsets(consumer confluentConsumer, partition kafka.TopicPartition) error {
	if partition.Error != nil {
		return fmt.Errorf("error storing offsets:%w", partition.Error)
	}
	offsets := []kafka.TopicPartition{partition}
	offsets[0].Offset++
	if _, err := consumer.StoreOffsets(offsets); err != nil {
		return err
	}
	return nil
}
