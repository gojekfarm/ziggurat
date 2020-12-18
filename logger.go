package ziggurat

import (
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"strings"
	"time"
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

func (l *Log) Infof(format string, v ...interface{}) {
	l.logger.Info().Msgf(format, v...)
}

func (l *Log) Debugf(format string, v ...interface{}) {
	l.logger.Debug().Msgf(format, v...)
}

func (l *Log) Warnf(format string, v ...interface{}) {
	l.logger.Warn().Msgf(format, v...)
}

func (l *Log) Errorf(format string, v ...interface{}) {
	if len(v) > 0 && v[0] != nil {
		l.logger.Error().Msgf(format, v...)
	}
}

func (l *Log) Fatalf(format string, v ...interface{}) {
	if len(v) > 0 && v[0] != nil {
		l.logger.Fatal().Msgf(format, v...)
	}
}

func (l *Log) Warn(format string) {
	l.logger.Warn().Msg(format)
}

func (l *Log) Info(format string) {
	l.logger.Info().Msg(format)
}

func (l *Log) Debug(format string) {
	l.logger.Debug().Msg(format)
}

func (l *Log) Error(format string) {
	l.errLogger.Error().Msg(format)
}

func (l *Log) Fatal(format string) {
	l.logger.Fatal().Msg(format)
}

func NewLogger(level string) *Log {
	zerolog.SetGlobalLevel(logLevelMapping[level])
	consoleLogger := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.Out = os.Stderr
		w.NoColor = true
		w.TimeFormat = time.RFC3339
		w.FormatMessage = func(i interface{}) string {
			return fmt.Sprintf("[ziggurat] %s", i)
		}
		w.FormatLevel = func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("[%s]", i))
		}
		w.TimeFormat = time.RFC850
	})
	loggerInst := zerolog.New(consoleLogger).With().Timestamp().Logger()
	errLoggerInst := zerolog.New(consoleLogger).With().Timestamp().CallerWithSkipFrameCount(callerFrameSkipCount).Logger()
	return &Log{
		errLogger: errLoggerInst,
		logger:    loggerInst,
	}
}
