package logger

import (
	"os"

	"github.com/rs/zerolog"
)

const callerFrameSkipCount = 3

const LevelInfo = "info"
const Disabled = "disabled"
const LevelWarn = "warn"
const LevelDebug = "debug"
const LevelFatal = "fatal"
const LevelError = "error"

var logLevelMapping = map[string]zerolog.Level{
	"debug":    zerolog.DebugLevel,
	"info":     zerolog.InfoLevel,
	"warn":     zerolog.WarnLevel,
	"error":    zerolog.ErrorLevel,
	"fatal":    zerolog.FatalLevel,
	"disabled": zerolog.Disabled,
}

type JSONLogger struct {
	errLogger zerolog.Logger
	logger    zerolog.Logger
}

func appendFields(l *zerolog.Event, kvs []map[string]interface{}) *zerolog.Event {
	for _, m := range kvs {
		l.Fields(m)
	}
	return l
}

func (l *JSONLogger) Info(message string, kvs ...map[string]interface{}) {
	appendFields(l.logger.Info(), kvs).Msg(message)
}

func (l *JSONLogger) Debug(message string, kvs ...map[string]interface{}) {
	appendFields(l.logger.Debug(), kvs).Msg(message)
}

func (l *JSONLogger) Warn(message string, kvs ...map[string]interface{}) {
	appendFields(l.logger.Warn(), kvs).Msg(message)
}

func (l *JSONLogger) Error(message string, err error, kvs ...map[string]interface{}) {
	if err != nil {
		appendFields(l.errLogger.Err(err), kvs).Msg(message)
	}
}

func (l *JSONLogger) Fatal(message string, err error, kvs ...map[string]interface{}) {
	if err != nil {
		appendFields(l.errLogger.Fatal(), kvs).Msg(message)
	}
}

func NewJSONLogger(level string) *JSONLogger {
	loggerInst := zerolog.New(os.Stdout).With().Str("log-type", "ziggurat").Timestamp().Logger().Level(logLevelMapping[level])
	errLoggerInst := zerolog.New(os.Stderr).With().Str("log-type", "ziggurat").Timestamp().Logger().Level(logLevelMapping[level])
	return &JSONLogger{
		errLogger: errLoggerInst,
		logger:    loggerInst,
	}
}
