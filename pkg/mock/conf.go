package mock

import (
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
)

type ConfigStore struct {
	ConfigFunc         func() *basic.Config
	ParseFunc          func(options basic.CommandLineOptions)
	GetByKeyFunc       func(key string) interface{}
	ValidateFunc       func(rules map[string]func(c *basic.Config) error) error
	UnmarshalByKeyFunc func(key string, model interface{}) error
}

func NewConfigStore() *ConfigStore {
	return &ConfigStore{
		ConfigFunc: func() *basic.Config {
			return &basic.Config{}
		},
		ParseFunc: func(options basic.CommandLineOptions) {

		},
		GetByKeyFunc: func(key string) interface{} {
			return nil
		},
		ValidateFunc: func(rules map[string]func(c *basic.Config) error) error {
			return nil
		},
		UnmarshalByKeyFunc: func(key string, model interface{}) error {
			return nil
		},
	}
}

func (c *ConfigStore) Config() *basic.Config {
	return c.ConfigFunc()
}

func (c *ConfigStore) Parse(options basic.CommandLineOptions) {
	c.ParseFunc(options)
}

func (c *ConfigStore) GetByKey(key string) interface{} {
	return c.GetByKeyFunc(key)
}

func (c *ConfigStore) Validate(rules map[string]func(c *basic.Config) error) error {
	return c.ValidateFunc(rules)
}

func (c *ConfigStore) UnmarshalByKey(key string, model interface{}) error {
	return c.UnmarshalByKeyFunc(key, model)
}
