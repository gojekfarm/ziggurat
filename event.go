package ziggurat

const HeaderMessageType = "x-message-type"
const HeaderMessageRoute = "x-message-route"

type Event interface {
	Value() []byte
	Key() []byte
	Headers() map[string]string
}
