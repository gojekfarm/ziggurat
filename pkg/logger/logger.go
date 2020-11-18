package logger

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

var consoleLogger = zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
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

var loggerInst = zerolog.New(consoleLogger).With().Timestamp().Logger()
var errLoggerInst = zerolog.New(consoleLogger).With().Timestamp().CallerWithSkipFrameCount(callerFrameSkipCount).Logger()

func ConfigureLogger(logLevel string) {
	logLevelInt := logLevelMapping[logLevel]
	loggerInst.Level(logLevelInt)
	errLoggerInst.Level(logLevelInt)
}

var LogError = func(err error, msg string, args map[string]interface{}) {
	if err != nil {
		errLoggerInst.Err(err).Fields(args).Msg(msg)
	}
}

var LogFatal = func(err error, msg string, args map[string]interface{}) {
	if err != nil {
		errLoggerInst.Fatal().Fields(args).Err(err).Msg(msg)
	}
}

var LogInfo = func(msg string, args map[string]interface{}) {
	loggerInst.Info().Fields(args).Msg(msg)
}

var LogWarn = func(msg string, args map[string]interface{}) {
	loggerInst.Warn().Fields(args).Msg(msg)
}
