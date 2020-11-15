package vconf

import (
	"github.com/gojekfarm/ziggurat-go/pkg/ziggurat/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/ziggurat/zerror"
)

var ConfigRules = map[string]func(c *basic.Config) error{
	"streamRouteValidation": func(c *basic.Config) error {
		if len(c.StreamRouter) == 0 {
			return zerror.ErrStreamRouteValidation
		}
		return nil
	},
	"serviceNameValidation": func(c *basic.Config) error {
		if c.ServiceName == "" {
			return zerror.ErrServiceNameValidation
		}
		return nil
	},
}
