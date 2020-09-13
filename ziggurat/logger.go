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

func ConfigureLogger(logLevel string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logLevelInt := logLevelMapping[logLevel]
	zerolog.SetGlobalLevel(logLevelInt)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}
