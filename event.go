package ziggurat

import (
	"bytes"
	"io"
	"time"
)

// Event is a generic event
// ReceivedTimestamp holds the timestamp of the message when it was received
// ProducerTimestamp holds the timestamp of the message as given by the producer
// Path is the message path and can be used by routers to route message to the correct handler
// EventType is the type of event eg:- kafka,rabbitmq,redis
// Headers can contain additional metadata about the message
type Event struct {
	Headers           map[string]string
	Value             []byte
	Key               []byte
	Path              string
	ProducerTimestamp time.Time
	ReceivedTimestamp time.Time
	EventType         string
}

// ValueReader can be used as a byte reader to read the message value
func ValueReader(e Event) io.Reader {
	return bytes.NewReader(e.Value)
}

// KeyReader can be used as a byte reader to read the message key
func KeyReader(e Event) io.Reader {
	return bytes.NewReader(e.Key)
}
