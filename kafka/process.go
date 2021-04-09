package kafka

import (
	"context"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
)

func kafkaProcessor(msg *kafka.Message, route string, c *kafka.Consumer, h ziggurat.Handler, l ziggurat.StructuredLogger, ctx context.Context) {
	timestamp := msg.Timestamp.Unix()
	topic := *msg.TopicPartition.Topic
	partition := int(msg.TopicPartition.Partition)

	headers := map[string]string{
		ziggurat.HeaderMessageRoute: route,
		ziggurat.HeaderMessageType:  "kafka",
		"x-kafka-timestamp":         strconv.FormatInt(timestamp, 10),
		"x-kafka-topic":             topic,
		"x-kafka-partition":         strconv.Itoa(partition),
	}

	event := NewMessage(msg.Value, msg.Key, headers)

	kvs := map[string]interface{}{
		"kafka-topic": topic,
		"partition":   partition,
		"route":       route,
	}
	if handlerError := h.Handle(ctx, event); handlerError != nil {
		l.Error("error processing message", handlerError, kvs)
	} else {
		l.Info("processed message successfully", kvs)
	}
	err := storeOffsets(c, msg.TopicPartition)
	l.Error("error storing offsets: %v", err)
}
