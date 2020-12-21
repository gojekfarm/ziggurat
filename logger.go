package ziggurat

import (
	"github.com/rs/zerolog"
	"os"
)

const callerFrameSkipCount = 3

var logLevelMapping = map[string]zerolog.Level{
	"debug":    zerolog.DebugLevel,
	"info":     zerolog.InfoLevel,
	"warn":     zerolog.WarnLevel,
	"error":    zerolog.ErrorLevel,
	"fatal":    zerolog.FatalLevel,
	"disabled": zerolog.Disabled,
}

type ZiggLogger struct {
	errLogger zerolog.Logger
	logger    zerolog.Logger
}

func appendFields(l *zerolog.Event, kvs []map[string]interface{}) *zerolog.Event {
	for _, m := range kvs {
		l.Fields(m)
	}
	return l
}

func (l *ZiggLogger) Info(message string, kvs ...map[string]interface{}) {
	appendFields(l.logger.Info(), kvs).Msg(message)
}

func (l *ZiggLogger) Debug(message string, kvs ...map[string]interface{}) {
	appendFields(l.logger.Debug(), kvs).Msg(message)
}

func (l *ZiggLogger) Warn(message string, kvs ...map[string]interface{}) {
	appendFields(l.logger.Warn(), kvs).Msg(message)
}

func (l *ZiggLogger) Error(message string, err error, kvs ...map[string]interface{}) {
	if err != nil {
		appendFields(l.errLogger.Err(err), kvs).Msg(message)
	}
}

func (l *ZiggLogger) Fatal(message string, err error, kvs ...map[string]interface{}) {
	if err != nil {
		appendFields(l.errLogger.Fatal(), kvs).Msg(message)
	}
}

func NewLogger(level string) *ZiggLogger {
	zerolog.SetGlobalLevel(logLevelMapping[level])
	loggerInst := zerolog.New(os.Stdout).With().Str("log-type", "ziggurat").Timestamp().Logger()
	errLoggerInst := zerolog.New(os.Stderr).With().Str("log-type", "ziggurat").Timestamp().CallerWithSkipFrameCount(callerFrameSkipCount).Logger()
	return &ZiggLogger{
		errLogger: errLoggerInst,
		logger:    loggerInst,
	}
}
