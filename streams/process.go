package streams

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
)

func kafkaProcessor(msg *kafka.Message, route string, c *kafka.Consumer, h ziggurat.Handler, l ziggurat.StructuredLogger, ctx context.Context) {
	headers := map[string]string{ziggurat.HeaderMessageRoute: route}
	event := ziggurat.CreateMessageEvent(msg.Value, headers, ctx)
	h.HandleEvent(event)
	err := storeOffsets(c, msg.TopicPartition)
	l.Error("error storing offsets: %v", err, nil)
}
