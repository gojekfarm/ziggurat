package zerror

import (
	"errors"
)

var (
	ErrNoHandlersRegistered     = errors.New("no handlers registered")
	ErrStreamRouteValidation    = errors.New("cfgReader validation error,stream-routes count is 0")
	ErrServiceNameValidation    = errors.New("cfgReader validation error, service-name is empty")
	ErrOffsetCommit             = errors.New("cannot commit errored message")
	ErrInvalidReturnCode        = errors.New("invalid processing function return code")
	ErrInvalidRouteRegistered   = errors.New("invalid route registered")
	ErrParsingStatsDConfig      = errors.New("error parsing statsd cfgReader")
	ErrRetryConsumerStopped     = errors.New("retry consumer stopped")
	ErrInterfaceNotProtoMessage = errors.New("interface must be of type proto.Message")
)

func returnErr(err error) error {
	if err != nil {
		return err
	}
	return nil
}
