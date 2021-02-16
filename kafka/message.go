package kafka

import "context"

type Message struct {
	value   []byte
	ctx     context.Context
	headers map[string]string
}

func (m Message) Value() []byte {
	return m.value
}

func (m Message) Headers() map[string]string {
	return m.headers
}

func CreateMessageEvent(value []byte, headers map[string]string) Message {
	return Message{
		value:   value,
		headers: headers,
	}
}
