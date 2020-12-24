package ziggurat

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func processor(msg *kafka.Message, route string, c *kafka.Consumer, h Handler, l StructuredLogger, ctx context.Context) {
	event := CreateMessageEvent(msg.Key, msg.Value, map[string]string{HeaderMessageType: "kafka", HeaderTypeRoute: route}, ctx)
	h.HandleMessage(event)
	err := storeOffsets(c, msg.TopicPartition)
	l.Error("error storing offsets: %v", err, nil)
}
