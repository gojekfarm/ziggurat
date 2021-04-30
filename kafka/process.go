package kafka

import (
	"context"
	"strconv"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
)

const (
	HeaderTopic     = "x-kafka-topic"
	HeaderPartition = "x-kafka-partition"
	EventType       = "kafka"
)

// processMessage executes the handler for every message that is received
// all metadata is serialized to strings and set in headers
func processMessage(ctx context.Context, msg *kafka.Message, c *kafka.Consumer, h ziggurat.Handler, l ziggurat.StructuredLogger, route string) {

	event := ziggurat.Event{
		Headers: map[string]string{
			HeaderPartition: strconv.Itoa(int(msg.TopicPartition.Partition)),
			HeaderTopic:     *msg.TopicPartition.Topic,
		},
		Value:             msg.Value,
		Key:               msg.Key,
		Path:              route,
		ProducerTimestamp: msg.Timestamp,
		ReceivedTimestamp: time.Now(),
		EventType:         EventType,
	}

	l.Error("kafka processing error", h.Handle(ctx, &event))
	err := storeOffsets(c, msg.TopicPartition)
	l.Error("error storing offsets: %v", err)
}
