package ziggurat

import (
	"time"
)

// Event is a generic event
// ReceivedTimestamp holds the timestamp of the message when it was received
// ProducerTimestamp holds the timestamp of the message as given by the producer
// Path is the message path and can be used by routers to route message to the correct handler
// EventType is the type of event eg:- kafka,rabbitmq,redis
// Headers can contain additional metadata about the message
type Event struct {
	Headers  map[string]string      `json:"headers"`
	Metadata map[string]interface{} `json:"meta"`
	Value    []byte                 `json:"value"`
	Key      []byte                 `json:"key"`
	// Path can be a single key
	Path string `json:"path"`
	// RoutingPath can be an actual path like string separated by a delimiter, this was added to maintain
	// backwards compatibility
	RoutingPath       string    `json:"routing_path"`
	ProducerTimestamp time.Time `json:"producer_timestamp"`
	ReceivedTimestamp time.Time `json:"received_timestamp"`
	EventType         string    `json:"event_type"`
}
