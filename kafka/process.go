package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gojekfarm/ziggurat"
)

const (
	EventType = "kafka"
)

func constructPath(rg string, topic string, part int32) string {
	return fmt.Sprintf("%s/%s/%d", rg, topic, part)
}

func processMessage(ctx context.Context, msg *kafka.Message, h ziggurat.Handler, l ziggurat.StructuredLogger, route string) {
	//copy kvs into new slices
	key := make([]byte, len(msg.Key))
	value := make([]byte, len(msg.Value))

	copy(key, msg.Key)
	copy(value, msg.Value)

	event := ziggurat.Event{
		Value: value,
		Key:   key,
		Metadata: map[string]interface{}{
			"kafka-topic":     *msg.TopicPartition.Topic,
			"kafka-partition": int(msg.TopicPartition.Partition),
		},
		RoutingPath:       constructPath(route, *msg.TopicPartition.Topic, msg.TopicPartition.Partition),
		ProducerTimestamp: msg.Timestamp,
		ReceivedTimestamp: time.Now(),
		EventType:         EventType,
	}
	l.Error("kafka processing error", h.Handle(ctx, &event))

}
