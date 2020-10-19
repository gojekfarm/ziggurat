package zig

import (
	"errors"
)

var (
	ErrNoHandlersRegistered     = errors.New("error: no handlers registered")
	ErrStreamRouteValidation    = errors.New("config validation error,stream-routes count is 0")
	ErrServiceNameValidation    = errors.New("config validation error, service-name is empty")
	ErrTopicEntityMismatch      = errors.New("topic entity mismatch")
	ErrOffsetCommit             = errors.New("cannot commit errored message")
	ErrReplayCountZero          = errors.New("replay count is 0, cannot 0 messages")
	ErrInvalidReturnCode        = errors.New("invalid processing function return code")
	ErrInvalidRouteRegistered   = errors.New("invalid route registered")
	ErrParsingStatsDConfig      = errors.New("error parsing statsd config")
	ErrParsingRabbitMQConfig    = errors.New("error parsing rabbitmq config")
	ErrRetryConsumerStopped     = errors.New("retry consumer stopped")
	ErrInterfaceNotProtoMessage = errors.New("interface must be of type proto.Message")
)
