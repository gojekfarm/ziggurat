package ziggurat

import "context"

const HeaderMessageType = "x-message-type"
const HeaderMessageRoute = "x-message-route"

type Event interface {
	Value() []byte
	Headers() map[string]string
	Context() context.Context
}

type MockEvent struct {
	ValueFunc   func() []byte
	HeadersFunc func() map[string]string
	ContextFunc func() context.Context
}

func CreateMockEvent() MockEvent {
	return MockEvent{
		ValueFunc: func() []byte {
			return []byte{}
		},
		HeadersFunc: func() map[string]string {
			return map[string]string{}
		},
		ContextFunc: func() context.Context {
			return context.Background()
		},
	}
}

func (m MockEvent) Value() []byte {
	return m.ValueFunc()
}

func (m MockEvent) Headers() map[string]string {
	return m.HeadersFunc()
}

func (m MockEvent) Context() context.Context {
	return m.ContextFunc()
}
