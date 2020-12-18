package ziggurat

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func processor(msg *kafka.Message, route string, c *kafka.Consumer, h MessageHandler, ctx context.Context) {
	event := NewMessageEvent(msg.Key, msg.Value, *msg.TopicPartition.Topic, route, msg.TimestampType.String(), msg.Timestamp)
	h.HandleMessage(event, ctx)
	LogError(storeOffsets(c, msg.TopicPartition), "consumer error", nil)
}
