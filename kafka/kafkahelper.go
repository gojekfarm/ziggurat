package kafka

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
)

var createConsumer = func(consumerConfig *kafka.ConfigMap, l ziggurat.StructuredLogger, topics []string) *kafka.Consumer {
	consumer, err := kafka.NewConsumer(consumerConfig)
	if err != nil {
		panic(err)
	}
	subscribeErr := consumer.SubscribeTopics(topics, nil)
	if subscribeErr != nil {
		panic(subscribeErr)
	}
	return consumer
}

var storeOffsets = func(consumer *kafka.Consumer, partition kafka.TopicPartition) error {

	// at least once delivery
	// offsets are stored in memory and are later flushed by the auto-commit timer

	if partition.Error != nil {
		return ErrOffsetCommit
	}
	offsets := []kafka.TopicPartition{partition}
	offsets[0].Offset++
	if _, err := consumer.StoreOffsets(offsets); err != nil {
		return err
	}
	return nil
}

var readMessage = func(c *kafka.Consumer, pollTimeout time.Duration) (*kafka.Message, error) {
	return c.ReadMessage(pollTimeout)
}

func NewConsumerConfig(bootstrapServers string, groupID string) *kafka.ConfigMap {
	return &kafka.ConfigMap{
		"bootstrap.servers":        bootstrapServers,
		"group.id":                 groupID,
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  5000,
		"debug":                    "consumer,broker",
		"go.logs.channel.enable":   true,
		"enable.auto.offset.store": false,
		//disable for at-least once delivery
	}
}
