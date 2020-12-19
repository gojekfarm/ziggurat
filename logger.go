package ziggurat

import (
	"github.com/rs/zerolog"
	"os"
)

const callerFrameSkipCount = 3

var logLevelMapping = map[string]zerolog.Level{
	"off":      zerolog.NoLevel,
	"debug":    zerolog.DebugLevel,
	"info":     zerolog.InfoLevel,
	"warn":     zerolog.WarnLevel,
	"error":    zerolog.ErrorLevel,
	"fatal":    zerolog.FatalLevel,
	"panic":    zerolog.PanicLevel,
	"disabled": zerolog.Disabled,
}

type Log struct {
	errLogger zerolog.Logger
	logger    zerolog.Logger
}

func appendFields(l *zerolog.Event, kvs []map[string]interface{}) *zerolog.Event {
	for _, m := range kvs {
		l.Fields(m)
	}
	return l
}

func (l *Log) Info(message string, kvs ...map[string]interface{}) {
	appendFields(l.logger.Info(), kvs).Msg(message)
}

func (l *Log) Debug(message string, kvs ...map[string]interface{}) {
	appendFields(l.logger.Debug(), kvs).Msg(message)
}

func (l *Log) Warn(message string, kvs ...map[string]interface{}) {
	appendFields(l.logger.Warn(), kvs).Msg(message)
}

func (l *Log) Error(message string, err error, kvs ...map[string]interface{}) {
	if err != nil {
		appendFields(l.errLogger.Err(err), kvs).Msg(message)
	}
}

func (l *Log) Fatal(message string, err error, kvs ...map[string]interface{}) {
	if err != nil {
		appendFields(l.errLogger.Fatal(), kvs).Msg(message)
	}
}

func NewLogger(level string) *Log {
	zerolog.SetGlobalLevel(logLevelMapping[level])
	loggerInst := zerolog.New(os.Stdout).With().Timestamp().Logger()
	errLoggerInst := zerolog.New(os.Stderr).With().Timestamp().CallerWithSkipFrameCount(callerFrameSkipCount).Logger()
	return &Log{
		errLogger: errLoggerInst,
		logger:    loggerInst,
	}
}
