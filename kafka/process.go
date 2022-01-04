package kafka

import (
	"context"
	"fmt"
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

func constructPath(rg string, topic string, part int32) string {
	return fmt.Sprintf("%s/%s/%d", rg, topic, part)
}

// processMessage executes the handler for every message that is received
// all metadata is serialized to string and set in headers
func processMessage(ctx context.Context, msg *kafka.Message, h ziggurat.Handler, l ziggurat.StructuredLogger, route string) {

	//copy kvs into new slices
	key := make([]byte, len(msg.Key))
	value := make([]byte, len(msg.Value))

	copy(key, msg.Key)
	copy(value, msg.Value)

	event := ziggurat.Event{
		Headers: map[string]string{
			HeaderPartition: strconv.Itoa(int(msg.TopicPartition.Partition)),
			HeaderTopic:     *msg.TopicPartition.Topic,
		},
		Value: value,
		Key:   key,
		Path:  route,
		Metadata: map[string]interface{}{
			"kafka-topic":     *msg.TopicPartition.Topic,
			"kafka-partition": msg.TopicPartition.Partition,
		},
		RoutingPath:       constructPath(route, *msg.TopicPartition.Topic, msg.TopicPartition.Partition),
		ProducerTimestamp: msg.Timestamp,
		ReceivedTimestamp: time.Now(),
		EventType:         EventType,
	}
	l.Error("kafka processing error", h.Handle(ctx, &event))

}
