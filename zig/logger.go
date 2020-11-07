package zig

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"time"
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

var routerLogger zerolog.Logger
var consumerLogger zerolog.Logger
var serverLogger zerolog.Logger
var retrierLogger zerolog.Logger
var metricLogger zerolog.Logger

func configureLogger(logLevel string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logLevelInt := logLevelMapping[logLevel]
	zerolog.SetGlobalLevel(logLevelInt)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true, TimeFormat: time.RFC3339,
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("|%s|", i))
		}})
	routerLogger = log.With().Str("component", "router").Logger()
	consumerLogger = log.With().Str("component", "consumer").Logger()
	serverLogger = log.With().Str("component", "http-server").Logger()
	retrierLogger = log.With().Str("component", "messageRetry").Logger()
	metricLogger = log.With().Str("component", "metrics").Logger()
}
