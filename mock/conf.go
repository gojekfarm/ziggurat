package mock

import (
	"github.com/gojekfarm/ziggurat/zb"
)

type ConfigStore struct {
	ConfigFunc         func() *zb.Config
	ParseFunc          func(options zb.CommandLineOptions) error
	GetByKeyFunc       func(key string) interface{}
	ValidateFunc       func(rules map[string]func(c *zb.Config) error) error
	UnmarshalByKeyFunc func(key string, model interface{}) error
}

func NewConfigStore() *ConfigStore {
	return &ConfigStore{
		ConfigFunc: func() *zb.Config {
			return &zb.Config{}
		},
		ParseFunc: func(options zb.CommandLineOptions) error {
			return nil
		},
		GetByKeyFunc: func(key string) interface{} {
			return nil
		},
		ValidateFunc: func(rules map[string]func(c *zb.Config) error) error {
			return nil
		},
		UnmarshalByKeyFunc: func(key string, model interface{}) error {
			return nil
		},
	}
}

func (c *ConfigStore) Config() *zb.Config {
	return c.ConfigFunc()
}

func (c *ConfigStore) Parse(options zb.CommandLineOptions) error {
	return c.ParseFunc(options)
}

func (c *ConfigStore) GetByKey(key string) interface{} {
	return c.GetByKeyFunc(key)
}

func (c *ConfigStore) Validate(rules map[string]func(c *zb.Config) error) error {
	return c.ValidateFunc(rules)
}

func (c *ConfigStore) UnmarshalByKey(key string, model interface{}) error {
	return c.UnmarshalByKeyFunc(key, model)
}
