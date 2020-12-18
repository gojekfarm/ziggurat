package ziggurat

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func processor(msg *kafka.Message, route string, c *kafka.Consumer, h MessageHandler, l LeveledLogger, ctx context.Context) {
	event := NewMessageEvent(msg.Key, msg.Value, *msg.TopicPartition.Topic, route, msg.TimestampType.String(), msg.Timestamp)
	h.HandleMessage(event, ctx)
	err := storeOffsets(c, msg.TopicPartition)
	l.Errorf("error storing offsets: %v", err)
}
