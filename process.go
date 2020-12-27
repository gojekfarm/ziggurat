package ziggurat

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func kafkaProcessor(msg *kafka.Message, route string, c *kafka.Consumer, h Handler, l StructuredLogger, ctx context.Context) {
	headers := map[string]string{HeaderMessageRoute: route}
	event := CreateMessageEvent(msg.Value, headers, ctx)
	h.HandleEvent(event)
	err := storeOffsets(c, msg.TopicPartition)
	l.Error("error storing offsets: %v", err, nil)
}
