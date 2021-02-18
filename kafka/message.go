package kafka

type Message struct {
	value   []byte
	headers map[string]string
}

func (m Message) Value() []byte {
	return m.value
}

func (m Message) Headers() map[string]string {
	return m.headers
}

func NewMessage(value []byte, headers map[string]string) Message {
	return Message{
		value:   value,
		headers: headers,
	}
}
