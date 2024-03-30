package logger

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

type TextLogger struct {
	l zerolog.Logger
}

func (h *TextLogger) Info(message string, kvs ...map[string]interface{}) {
	appendFields(h.l.Info(), kvs).Msg(message)
}

func (h *TextLogger) Debug(message string, kvs ...map[string]interface{}) {
	appendFields(h.l.Debug(), kvs).Msg(message)
}

func (h *TextLogger) Warn(message string, kvs ...map[string]interface{}) {
	appendFields(h.l.Warn(), kvs).Msg(message)
}

func (h *TextLogger) Error(message string, err error, kvs ...map[string]interface{}) {
	if err != nil {
		appendFields(h.l.Err(err), kvs).Msg(message)
	}
}

func (h *TextLogger) Fatal(message string, err error, kvs ...map[string]interface{}) {
	if err != nil {
		appendFields(h.l.Fatal(), kvs).Msg(message)
	}
}

func NewLogger(level string, opts ...func(w *zerolog.ConsoleWriter)) *TextLogger {
	cw := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.FormatLevel = func(i interface{}) string {
			return fmt.Sprintf("[%s]", i)
		}
		w.TimeFormat = time.RFC822
		w.NoColor = true

		for _, o := range opts {
			o(w)
		}
	})
	l := zerolog.New(cw).With().Timestamp().Logger().Level(logLevelMapping[level])
	return &TextLogger{l: l}
}
