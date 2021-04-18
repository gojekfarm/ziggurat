package kafka

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
	"strconv"
	"time"
)

const (
	HeaderTopic     = "x-kafka-topic"
	HeaderPartition = "x-kafka-partition"
)

// processMessage executed the handler for every message that is received
// all metadata is serialized to strings and set in headers
func processMessage(msg *kafka.Message, route string, c *kafka.Consumer, h ziggurat.Handler, l ziggurat.StructuredLogger, ctx context.Context) {

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
		EventType:         "kafka",
	}

	h.Handle(ctx, &event)
	err := storeOffsets(c, msg.TopicPartition)
	l.Error("error storing offsets: %v", err)
}
