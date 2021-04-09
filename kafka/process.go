package kafka

import (
	"context"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
)

const (
	HeaderTimestamp = "x-kafka-timestamp"
	HeaderTopic     = "x-kafka-topic"
	HeaderPartition = "x-kafka-partition"
)

func kafkaProcessor(msg *kafka.Message, route string, c *kafka.Consumer, h ziggurat.Handler, l ziggurat.StructuredLogger, ctx context.Context) {
	timestamp := msg.Timestamp.Unix()
	topic := *msg.TopicPartition.Topic
	partition := int(msg.TopicPartition.Partition)

	headers := map[string]string{
		ziggurat.HeaderMessageRoute: route,
		ziggurat.HeaderMessageType:  "kafka",
		HeaderTimestamp:             strconv.FormatInt(timestamp, 10),
		HeaderTopic:                 topic,
		HeaderPartition:             strconv.Itoa(partition),
	}

	event := NewMessage(msg.Value, msg.Key, headers)

	h.Handle(ctx, event)
	err := storeOffsets(c, msg.TopicPartition)
	l.Error("error storing offsets: %v", err)
}
