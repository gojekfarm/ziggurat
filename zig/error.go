package zig

import (
	"errors"
)

var (
	ErrNoHandlersRegistered     = errors.New("error: no handlers registered")
	ErrStreamRouteValidation    = errors.New("config validation error,stream-routes count is 0")
	ErrServiceNameValidation    = errors.New("config validation error, service-name is empty")
	ErrOffsetCommit             = errors.New("cannot commit errored message")
	ErrInvalidReturnCode        = errors.New("invalid processing function return code")
	ErrInvalidRouteRegistered   = errors.New("invalid route registered")
	ErrParsingStatsDConfig      = errors.New("error parsing statsd config")
	ErrRetryConsumerStopped     = errors.New("retry consumer stopped")
	ErrInterfaceNotProtoMessage = errors.New("interface must be of type proto.Message")
)
