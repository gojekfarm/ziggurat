package kafka

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gojekfarm/ziggurat"
)

var createConsumer = func(consumerConfig *kafka.ConfigMap, l ziggurat.StructuredLogger, topics []string) *kafka.Consumer {
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

var closeConsumer = func(c *kafka.Consumer) error {
	return c.Close()
}

// storeOffsets ensures at least once delivery
// offsets are stored in memory and are later flushed by the auto-commit timer
var storeOffsets = func(consumer *kafka.Consumer, partition kafka.TopicPartition) error {
	if partition.Error != nil {
		return ErrPart
	}
	offsets := []kafka.TopicPartition{partition}
	offsets[0].Offset++
	if _, err := consumer.StoreOffsets(offsets); err != nil {
		return err
	}
	return nil
}

var pollEvent = func(c *kafka.Consumer, pollTimeout int) kafka.Event {
	return c.Poll(pollTimeout)
}
