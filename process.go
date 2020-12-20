package ziggurat

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func processor(msg *kafka.Message, route string, c *kafka.Consumer, h Handler, l StructuredLogger, ctx context.Context) {
	attributes := MsgAttributes{
		"kafka-timestamp": msg.Timestamp,
		"kafka-partition": msg.TopicPartition.Partition,
		"kafka-topic":     *msg.TopicPartition.Topic,
	}
	event := CreateMessage(msg.Key, msg.Value, route, attributes)
	h.HandleMessage(event, ctx)
	err := storeOffsets(c, msg.TopicPartition)
	l.Error("error storing offsets: %v", err, nil)
}
