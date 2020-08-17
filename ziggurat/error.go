package ziggurat

import "errors"

var (
	ErrNoHandlersRegistered  = errors.New("error: no handlers registered")
	ErrStreamRouteValidation = errors.New("config validation error,stream-routes count is 0")
	ErrServiceNameValidation = errors.New("config validation error, service-name is empty")
)
