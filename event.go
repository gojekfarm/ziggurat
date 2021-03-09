package ziggurat

const HeaderMessageType = "x-message-type"
const HeaderMessageRoute = "x-message-route"

type Event interface {
	Value() []byte
	Key() []byte
	Headers() map[string]string
}

type MockEvent struct {
	ValueFunc   func() []byte
	HeadersFunc func() map[string]string
	KeyFunc     func() []byte
}

func CreateMockEvent() MockEvent {
	return MockEvent{
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

func (m MockEvent) Value() []byte {
	return m.ValueFunc()
}

func (m MockEvent) Headers() map[string]string {
	return m.HeadersFunc()
}

func (m MockEvent) Key() []byte {
	return m.KeyFunc()
}
