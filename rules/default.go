package rules

import (
	"github.com/gojekfarm/ziggurat/zb"
	"github.com/gojekfarm/ziggurat/zerror"
)

var DefaultRules = map[string]func(c *zb.Config) error{
	"streamRouteValidation": func(c *zb.Config) error {
		if len(c.StreamRouter) == 0 {
			return zerror.ErrStreamRouteValidation
		}
		return nil
	},
	"serviceNameValidation": func(c *zb.Config) error {
		if c.ServiceName == "" {
			return zerror.ErrServiceNameValidation
		}
		return nil
	},
}
