package ziggurat

const HeaderMessageType = "x-message-type"
const HeaderMessageRoute = "x-message-route"

type Event interface {
	Value() []byte
	Headers() map[string]string
}

type MockEvent struct {
	ValueFunc   func() []byte
	HeadersFunc func() map[string]string
}

func CreateMockEvent() MockEvent {
	return MockEvent{
		ValueFunc: func() []byte {
			return []byte{}
		},
		HeadersFunc: func() map[string]string {
			return map[string]string{}
		},
	}
}

func (m MockEvent) Value() []byte {
	return m.ValueFunc()
}

func (m MockEvent) Headers() map[string]string {
	return m.HeadersFunc()
}
