package ziggurat

// HeaderMessageType is an event header which specifies the type of message being received
const HeaderMessageType = "x-message-type"

// HeaderMessageRoute is an event header which can be used by third-party handlers to route messages
const HeaderMessageRoute = "x-message-route"

// Event is used to represent a generic event produced by streams
// Messages produced by your streams must implement this interface
type Event interface {
	Value() []byte
	Key() []byte
	Headers() map[string]string
}
