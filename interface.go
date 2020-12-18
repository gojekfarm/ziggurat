package ziggurat

import "context"

type MessageHandler interface {
	HandleMessage(event MessageEvent, ctx context.Context) ProcessStatus
}

type Streams interface {
	Consume(ctx context.Context, routes Routes, handler MessageHandler) chan error
}

type LeveledLogger interface {
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
	Warn(format string)
	Info(format string)
	Debug(format string)
	Error(format string)
	Fatal(format string)
}