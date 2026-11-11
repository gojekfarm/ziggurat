// Package eventmeta extracts Kafka routing facts from a ziggurat.Event for the
// metrics, tracing and logging middlewares, so they all agree on how a routing path
// (<group>/<topic>/<partition>) is split.
package eventmeta

import (
	"fmt"
	"strings"

	"github.com/gojekfarm/ziggurat/v2"
)

// Kafka returns the consumer group, topic and partition of an event produced by
// kafka.ConsumerGroup. ok is false when the event carries no Kafka topic/partition
// metadata, or when its routing path does not end in "/<topic>/<partition>" (e.g. a
// custom path) - in which case nothing can be said reliably and callers fall back to
// the raw routing path.
//
// The partition is accepted as int or float64 (see asInt).
//
// The group is everything before that suffix, so group IDs that contain "/" work.
func Kafka(event *ziggurat.Event) (group, topic string, partition int, ok bool) {
	topic, _ = event.Metadata[ziggurat.MetadataKeyKafkaTopic].(string)
	partition, hasPartition := asInt(event.Metadata[ziggurat.MetadataKeyKafkaPartition])
	if topic == "" || !hasPartition {
		return "", "", 0, false
	}
	suffix := fmt.Sprintf("/%s/%d", topic, partition)
	if !strings.HasSuffix(event.RoutingPath, suffix) {
		return "", "", 0, false
	}
	return strings.TrimSuffix(event.RoutingPath, suffix), topic, partition, true
}

// RabbitMQQueue returns the queue a message was just delivered from, if the event was
// consumed through mw/rabbitmq.
func RabbitMQQueue(event *ziggurat.Event) (string, bool) {
	q, ok := event.Metadata[ziggurat.MetadataKeyRabbitMQQueue].(string)
	return q, ok && q != ""
}

// asInt reads a metadata number. The Kafka consumer stores an int, but an event that
// went through RabbitMQ (retry/DLQ) was JSON round-tripped, which turns it into float64.
func asInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int32:
		return int(n), true
	case float64:
		return int(n), n == float64(int(n))
	}
	return 0, false
}
