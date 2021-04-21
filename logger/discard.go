package logger

type DiscardLogger struct {
	InfoFunc  func(message string, kvs ...map[string]interface{})
	DebugFunc func(message string, kvs ...map[string]interface{})
	WarnFunc  func(message string, kvs ...map[string]interface{})
	ErrorFunc func(message string, err error, kvs ...map[string]interface{})
	FatalFunc func(message string, err error, kvs ...map[string]interface{})
}

func (d DiscardLogger) Info(message string, kvs ...map[string]interface{}) {
	d.InfoFunc(message, kvs...)
}

func (d DiscardLogger) Debug(message string, kvs ...map[string]interface{}) {
	d.DebugFunc(message, kvs...)
}

func (d DiscardLogger) Warn(message string, kvs ...map[string]interface{}) {
	d.WarnFunc(message, kvs...)
}

func (d DiscardLogger) Error(message string, err error, kvs ...map[string]interface{}) {
	d.ErrorFunc(message, err, kvs...)
}

func (d DiscardLogger) Fatal(message string, err error, kvs ...map[string]interface{}) {
	d.FatalFunc(message, err, kvs...)
}

func NewDiscardLogger() *DiscardLogger {
	return &DiscardLogger{
		InfoFunc: func(message string, kvs ...map[string]interface{}) {

		},
		DebugFunc: func(message string, kvs ...map[string]interface{}) {

		},
		WarnFunc: func(message string, kvs ...map[string]interface{}) {

		},
		ErrorFunc: func(message string, err error, kvs ...map[string]interface{}) {

		},
		FatalFunc: func(message string, err error, kvs ...map[string]interface{}) {

		},
	}
}
