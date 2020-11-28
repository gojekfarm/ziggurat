package rules

import (
	"github.com/gojekfarm/ziggurat-go/pkg/zbasic"
	"github.com/gojekfarm/ziggurat-go/pkg/zerror"
)

var DefaultRules = map[string]func(c *zbasic.Config) error{
	"streamRouteValidation": func(c *zbasic.Config) error {
		if len(c.StreamRouter) == 0 {
			return zerror.ErrStreamRouteValidation
		}
		return nil
	},
	"serviceNameValidation": func(c *zbasic.Config) error {
		if c.ServiceName == "" {
			return zerror.ErrServiceNameValidation
		}
		return nil
	},
}
