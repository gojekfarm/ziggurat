package kafka

import "time"

type Message struct {
	value     []byte
	key       []byte
	headers   map[string]string
	Timestamp time.Time
	Partition int
	Topic     string
}

func (m Message) Value() []byte {
	return m.value
}

func (m Message) Headers() map[string]string {
	return m.headers
}

func (m Message) Key() []byte {
	return m.key
}

func NewMessage(value []byte, key []byte, headers map[string]string) Message {
	return Message{
		value:   value,
		headers: headers,
	}
}
