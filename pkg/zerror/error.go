package zerror

import (
	"errors"
)

var (
	ErrNoHandlersRegistered  = errors.New("no handlers registered")
	ErrStreamRouteValidation = errors.New("cfgReader validation error,stream-routes count is 0")
	ErrServiceNameValidation = errors.New("cfgReader validation error, service-name is empty")
	ErrOffsetCommit          = errors.New("cannot commit errored message")
	ErrNoDecoderFound        = errors.New("no decoder found")
)
