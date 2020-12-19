package ziggurat

import (
	"context"
)

type MockKStreams struct {
	ConsumeFunc func(ctx context.Context, routes Routes, handler Handler) chan error
}

func (m MockKStreams) Consume(ctx context.Context, routes Routes, handler Handler) chan error {
	return m.ConsumeFunc(ctx, routes, handler)
}

func NewMockKafkaStreams() *MockKStreams {
	return &MockKStreams{
		ConsumeFunc: func(ctx context.Context, routes Routes, handler Handler) chan error {
			return make(chan error)
		},
	}
}

//type MockLeveledLogger struct {
//	InfofFunc  func(format string, v ...interface{})
//	ErrorfFunc func(format string, v ...interface{})
//	WarnfFunc  func(format string, v ...interface{})
//	FatalfFunc func(format string, v ...interface{})
//	DebugfFunc func(format string, v ...interface{})
//}
//
//func (m MockLeveledLogger) Info(message string, kvs map[string]interface{}) {
//	m.InfofFunc(message, kvs...)
//}
//
//func (m MockLeveledLogger) Debug(message string, kvs map[string]interface{}) {
//	m.DebugfFunc(message, kvs...)
//}
//
//func (m MockLeveledLogger) Warn(message string, kvs map[string]interface{}) {
//	m.WarnfFunc(message, kvs...)
//}
//
//func (m MockLeveledLogger) Error(message string, err error, kvs map[string]interface{}) {
//	m.ErrorfFunc(message, v...)
//}
//
//func (m MockLeveledLogger) Fatal(message string, err error, kvs map[string]interface{}) {
//	m.FatalfFunc(format, kvs...)
//}
//
//func NewMockLogger(level string) *MockLeveledLogger {
//	return &MockLeveledLogger{
//		InfofFunc: func(format string, v ...interface{}) {
//
//		},
//		ErrorfFunc: func(format string, v ...interface{}) {
//
//		},
//		WarnfFunc: func(format string, v ...interface{}) {
//
//		},
//		FatalfFunc: func(format string, v ...interface{}) {
//
//		},
//		DebugfFunc: func(format string, v ...interface{}) {
//
//		},
//	}
//}
