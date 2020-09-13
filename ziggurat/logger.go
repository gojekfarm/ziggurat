package ziggurat

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

var logLevelMapping = map[string]zerolog.Level{
	"off":   zerolog.NoLevel,
	"debug": zerolog.DebugLevel,
	"info":  zerolog.InfoLevel,
	"warn":  zerolog.WarnLevel,
	"error": zerolog.ErrorLevel,
	"fatal": zerolog.FatalLevel,
	"panic": zerolog.PanicLevel,
}

var RouterLogger zerolog.Logger
var ConsumerLogger zerolog.Logger
var ServerLogger zerolog.Logger
var RetrierLogger zerolog.Logger

func ConfigureLogger(logLevel string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logLevelInt := logLevelMapping[logLevel]
	zerolog.SetGlobalLevel(logLevelInt)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	RouterLogger = log.With().Str("component", "router").Logger()
	ConsumerLogger = log.With().Str("component", "consumer").Logger()
	ServerLogger = log.With().Str("component", "http-server").Logger()
	RetrierLogger = log.With().Str("component", "retrier").Logger()
}
