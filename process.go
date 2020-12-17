package ziggurat

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func processor(msg *kafka.Message, route string, c *kafka.Consumer, h MessageHandler, a AppContext) {
	event := NewMessageEvent(msg.Key, msg.Value, *msg.TopicPartition.Topic, route, msg.TimestampType.String(), msg.Timestamp)
	h.HandleMessage(event, a)
	LogError(storeOffsets(c, msg.TopicPartition), "consumer error", nil)
}
