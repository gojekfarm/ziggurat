package ziggurat

import (
	"context"
)

type MockKStreams struct {
	ConsumeFunc func(ctx context.Context, handler Handler) chan error
}

func (m MockKStreams) Stream(ctx context.Context, handler Handler) chan error {
	return m.ConsumeFunc(ctx, handler)
}

func NewMockKafkaStreams() *MockKStreams {
	return &MockKStreams{
		ConsumeFunc: func(ctx context.Context, handler Handler) chan error {
			return make(chan error)
		},
	}
}

type MockStructureLogger struct {
	InfoFunc  func(m string, kv ...map[string]interface{})
	WarnFunc  func(m string, kv ...map[string]interface{})
	ErrorFunc func(m string, kv ...map[string]interface{})
	FatalFunc func(m string, kv ...map[string]interface{})
	DebugFunc func(m string, kv ...map[string]interface{})
}

func (m MockStructureLogger) Info(message string, kvs ...map[string]interface{}) {
	m.InfoFunc(message, kvs...)
}

func (m MockStructureLogger) Debug(message string, kvs ...map[string]interface{}) {
	m.DebugFunc(message, kvs...)
}

func (m MockStructureLogger) Warn(message string, kvs ...map[string]interface{}) {
	m.WarnFunc(message, kvs...)
}

func (m MockStructureLogger) Error(message string, err error, kvs ...map[string]interface{}) {
	m.ErrorFunc(message, kvs...)
}

func (m MockStructureLogger) Fatal(message string, err error, kvs ...map[string]interface{}) {
	m.FatalFunc(message, kvs...)
}
