package ziggurat

import "context"

const HeaderMessageType = "x-message-type"
const HeaderMessageRoute = "x-message-route"

type Event interface {
	Value() []byte
	Headers() map[string]string
	Context() context.Context
}
