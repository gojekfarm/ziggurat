package kafka

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
)

var createConsumerMock = func(consumerConfig *kafka.ConfigMap, l ziggurat.StructuredLogger, topics []string) *kafka.Consumer {
	return &kafka.Consumer{}
}

var closeConsumerMock = func(c *kafka.Consumer) error {
	return nil
}

var pollEventMock = func(c *kafka.Consumer, pollTimeout int) kafka.Event {
	topic := "foo"
	return &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: 0},
	}
}

var storeOffsetsMock = func(consumer *kafka.Consumer, partition kafka.TopicPartition) error {
	return nil
}
