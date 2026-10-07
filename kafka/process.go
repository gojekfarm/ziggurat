package kafka

import (
	"context"
	"fmt"
	"github.com/gojekfarm/ziggurat/v2"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	EventType = "kafka"
)

func constructPath(rg string, topic string, part int32) string {
	return fmt.Sprintf("%s/%s/%d", rg, topic, part)
}

// headersToMap converts Kafka's native []kafka.Header into a map[string]string, suitable
// for Event.Metadata[ziggurat.MetadataKeyHeaders]. Kafka technically allows duplicate
// header keys; last one wins here, which is sufficient for the single-valued headers
// (e.g. a W3C traceparent) this is meant to carry.
func headersToMap(headers []kafka.Header) map[string]string {
	if len(headers) == 0 {
		return nil
	}
	m := make(map[string]string, len(headers))
	for _, h := range headers {
		m[h.Key] = string(h.Value)
	}
	return m
}

func processMessage(ctx context.Context, msg *kafka.Message, h ziggurat.Handler, route string) {
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
	// Only set the headers key when there's something to set: storing a nil
	// map[string]string literally into the key (rather than omitting it) would wrap a
	// nil map in a non-nil interface{} value, which is a classic Go footgun for anyone
	// checking Metadata[MetadataKeyHeaders] != nil rather than type-asserting first.
	if headers := headersToMap(msg.Headers); headers != nil {
		event.Metadata[ziggurat.MetadataKeyHeaders] = headers
	}
	h.Handle(ctx, &event)

}
