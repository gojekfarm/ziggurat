package ziggurat

import (
	"context"
)

type mockStreams struct {
	ConsumeFunc func(ctx context.Context, handler Handler) chan error
}

func (m mockStreams) Stream(ctx context.Context, handler Handler) chan error {
	return m.ConsumeFunc(ctx, handler)
}

type mockStructureLogger struct {
	InfoFunc  func(m string, kv ...map[string]interface{})
	WarnFunc  func(m string, kv ...map[string]interface{})
	ErrorFunc func(m string, kv ...map[string]interface{})
	FatalFunc func(m string, kv ...map[string]interface{})
	DebugFunc func(m string, kv ...map[string]interface{})
}

func (m mockStructureLogger) Info(message string, kvs ...map[string]interface{}) {
	m.InfoFunc(message, kvs...)
}

func (m mockStructureLogger) Debug(message string, kvs ...map[string]interface{}) {
	m.DebugFunc(message, kvs...)
}

func (m mockStructureLogger) Warn(message string, kvs ...map[string]interface{}) {
	m.WarnFunc(message, kvs...)
}

func (m mockStructureLogger) Error(message string, err error, kvs ...map[string]interface{}) {
	m.ErrorFunc(message, kvs...)
}

func (m mockStructureLogger) Fatal(message string, err error, kvs ...map[string]interface{}) {
	m.FatalFunc(message, kvs...)
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
