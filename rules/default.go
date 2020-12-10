package rules

import (
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/zerror"
)

var DefaultRules = map[string]func(c *zbase.Config) error{
	"streamRouteValidation": func(c *zbase.Config) error {
		if len(c.StreamRouter) == 0 {
			return zerror.ErrStreamRouteValidation
		}
		return nil
	},
	"serviceNameValidation": func(c *zbase.Config) error {
		if c.ServiceName == "" {
			return zerror.ErrServiceNameValidation
		}
		return nil
	},
}
