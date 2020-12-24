package ziggurat

import (
	"context"
)

const HeaderMessageType = "x-message-type"
const HeaderTypeRoute = "x-message-route"

type Message struct {
	value          []byte
	key            []byte
	ctx            context.Context
	MessageHeaders map[string]string
}

func (m Message) Value() []byte {
	return m.value
}

func (m Message) Key() []byte {
	return m.key
}

func (m Message) Context() context.Context {
	return m.ctx
}

func (m Message) Header(key string) string {
	return m.MessageHeaders[key]
}

func CreateMessageEvent(key []byte, value []byte, headers map[string]string, ctx context.Context) Message {
	return Message{
		value:          value,
		key:            key,
		ctx:            ctx,
		MessageHeaders: headers,
	}
}
