package rabbitmq

import (
	"fmt"
	"github.com/gojekfarm/ziggurat"
)

type amqpExtraLogger struct {
	l ziggurat.StructuredLogger
}

func (a *amqpExtraLogger) Printf(format string, v ...interface{}) {
	s := fmt.Sprintf("[AMQPEXTRA] " + format, v...)
	a.l.Info(s)
}
