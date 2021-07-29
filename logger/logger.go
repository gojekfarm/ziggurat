package logger

import (
	"fmt"

	"github.com/rs/zerolog"
)

type humanLogger struct {
	l zerolog.Logger
}

func (h *humanLogger) Info(message string, kvs ...map[string]interface{}) {
	appendFields(h.l.Info(), kvs).Msg(message)
}

func (h *humanLogger) Debug(message string, kvs ...map[string]interface{}) {
	appendFields(h.l.Debug(), kvs).Msg(message)
}

func (h *humanLogger) Warn(message string, kvs ...map[string]interface{}) {
	appendFields(h.l.Warn(), kvs).Msg(message)
}

func (h *humanLogger) Error(message string, err error, kvs ...map[string]interface{}) {
	if err != nil {
		appendFields(h.l.Err(err), kvs).Msg(message)
	}
}

func (h *humanLogger) Fatal(message string, err error, kvs ...map[string]interface{}) {
	if err != nil {
		appendFields(h.l.Fatal(), kvs).Msg(message)
	}
}

func NewLogger(level string, opts ...func(w *zerolog.ConsoleWriter)) *humanLogger {
	cw := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.FormatLevel = func(i interface{}) string {
			return fmt.Sprintf("[%s]", i)
		}

		w.NoColor = true

		for _, o := range opts {
			o(w)
		}
	})
	l := zerolog.New(cw).With().Timestamp().Logger().Level(logLevelMapping[level])
	return &humanLogger{l: l}
}
