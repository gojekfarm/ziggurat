package mock

import (
	"github.com/gojekfarm/ziggurat/zbase"
)

type ConfigStore struct {
	ConfigFunc         func() *zbase.Config
	ParseFunc          func(options zbase.CommandLineOptions) error
	GetByKeyFunc       func(key string) interface{}
	ValidateFunc       func(rules map[string]func(c *zbase.Config) error) error
	UnmarshalByKeyFunc func(key string, model interface{}) error
}

func NewConfigStore() *ConfigStore {
	return &ConfigStore{
		ConfigFunc: func() *zbase.Config {
			return &zbase.Config{}
		},
		ParseFunc: func(options zbase.CommandLineOptions) error {
			return nil
		},
		GetByKeyFunc: func(key string) interface{} {
			return nil
		},
		ValidateFunc: func(rules map[string]func(c *zbase.Config) error) error {
			return nil
		},
		UnmarshalByKeyFunc: func(key string, model interface{}) error {
			return nil
		},
	}
}

func (c *ConfigStore) Config() *zbase.Config {
	return c.ConfigFunc()
}

func (c *ConfigStore) Parse(options zbase.CommandLineOptions) error {
	return c.ParseFunc(options)
}

func (c *ConfigStore) GetByKey(key string) interface{} {
	return c.GetByKeyFunc(key)
}

func (c *ConfigStore) Validate(rules map[string]func(c *zbase.Config) error) error {
	return c.ValidateFunc(rules)
}

func (c *ConfigStore) UnmarshalByKey(key string, model interface{}) error {
	return c.UnmarshalByKeyFunc(key, model)
}
