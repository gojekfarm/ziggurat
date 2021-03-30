package mock

type Event struct {
	ValueFunc   func() []byte
	HeadersFunc func() map[string]string
	KeyFunc     func() []byte
}

func CreateMockEvent() Event {
	return Event{
		ValueFunc: func() []byte {
			return []byte{}
		},
		HeadersFunc: func() map[string]string {
			return map[string]string{}
		},
		KeyFunc: func() []byte {
			return []byte{}
		},
	}
}

func (m Event) Value() []byte {
	return m.ValueFunc()
}

func (m Event) Headers() map[string]string {
	return m.HeadersFunc()
}

func (m Event) Key() []byte {
	return m.KeyFunc()
}
