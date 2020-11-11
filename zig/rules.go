package zig

var configRules = map[string]func(c *Config) error{
	"streamRouteValidation": func(c *Config) error {
		if len(c.StreamRouter) == 0 {
			return ErrStreamRouteValidation
		}
		return nil
	},
	"serviceNameValidation": func(c *Config) error {
		if c.ServiceName == "" {
			return ErrServiceNameValidation
		}
		return nil
	},
}
