package ziggurat

import "time"

// MetadataKeyHeaders is the Event.Metadata key under which the native transport message
// headers (Kafka record headers, AMQP message headers) are captured, as map[string]string.
// This is the shared vehicle for anything that needs to ride along on a message's
// headers - most notably a W3C traceparent/tracestate for distributed tracing context
// propagation across consume/republish hops.
const MetadataKeyHeaders = "headers"

// Well-known Event.Metadata keys set by the bundled consumers. Middlewares (metrics,
// tracing, logging) read these instead of parsing RoutingPath.
//
// Note: a message republished through mw/rabbitmq is serialised with its Metadata, so a
// retried Kafka event still carries MetadataKeyKafkaTopic/Partition from its origin.
// mw/rabbitmq overwrites MetadataKeyRabbitMQQueue on every consume, so its presence
// means "delivered by RabbitMQ right now".
const (
	MetadataKeyKafkaTopic     = "kafka-topic"
	MetadataKeyKafkaPartition = "kafka-partition"
	MetadataKeyRabbitMQQueue  = "rabbitmq-queue"
)

// Event is a generic event
// ReceivedTimestamp holds the timestamp of the message when it was received
// ProducerTimestamp holds the timestamp of the message as given by the producer
// Path is the message path and can be used by routers to route message to the correct handler
// EventType is the type of event eg:- kafka,rabbitmq,redis
// Metadata is used to store metadata about the message
type Event struct {
	Metadata map[string]any `json:"meta"`
	Value    []byte         `json:"value"`
	Key      []byte         `json:"key"`
	// RoutingPath can be an actual path like a string separated by a delimiter
	RoutingPath       string    `json:"routing_path"`
	ProducerTimestamp time.Time `json:"producer_timestamp"`
	ReceivedTimestamp time.Time `json:"received_timestamp"`
	EventType         string    `json:"event_type"`
}
