package ziggurat

import (
	"errors"
	"github.com/rs/zerolog/log"
)

var (
	ErrNoHandlersRegistered  = errors.New("error: no handlers registered")
	ErrStreamRouteValidation = errors.New("config validation error,stream-routes count is 0")
	ErrServiceNameValidation = errors.New("config validation error, service-name is empty")
	ErrTopicEntityMismatch   = errors.New("topic entity mismatch")
	ErrOffsetCommit          = errors.New("cannot commit errored message")
	ErrReplayCountZero       = errors.New("replay count is 0, cannot 0 messages")
	ErrStatsdClientNil       = errors.New("statsd client is nil")
)

func logErr(err error, msg string) {
	if err != nil {
		log.Error().Err(err).Msg(msg)
	}
}
