package kafka

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
)

func kafkaProcessor(msg *kafka.Message, route string, c *kafka.Consumer, h ziggurat.Handler, l ziggurat.StructuredLogger, ctx context.Context) {
	headers := map[string]string{
		ziggurat.HeaderMessageRoute: route,
		ziggurat.HeaderMessageType:  "kafka",
	}
	event := NewMessage(msg.Value, headers)
	event.Timestamp = msg.Timestamp
	event.Topic = *msg.TopicPartition.Topic
	event.Partition = int(msg.TopicPartition.Partition)

	h.HandleEvent(ctx, event)
	err := storeOffsets(c, msg.TopicPartition)
	l.Error("error storing offsets: %v", err, map[string]interface{}{"topic": event.Topic, "partition": event.Partition})
}
