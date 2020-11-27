package mock

import (
	"github.com/gojekfarm/ziggurat-go/pkg/zbasic"
)

type ConfigStore struct {
	ConfigFunc         func() *zbasic.Config
	ParseFunc          func(options zbasic.CommandLineOptions)
	GetByKeyFunc       func(key string) interface{}
	ValidateFunc       func(rules map[string]func(c *zbasic.Config) error) error
	UnmarshalByKeyFunc func(key string, model interface{}) error
}

func NewConfigStore() *ConfigStore {
	return &ConfigStore{
		ConfigFunc: func() *zbasic.Config {
			return &zbasic.Config{}
		},
		ParseFunc: func(options zbasic.CommandLineOptions) {

		},
		GetByKeyFunc: func(key string) interface{} {
			return nil
		},
		ValidateFunc: func(rules map[string]func(c *zbasic.Config) error) error {
			return nil
		},
		UnmarshalByKeyFunc: func(key string, model interface{}) error {
			return nil
		},
	}
}

func (c *ConfigStore) Config() *zbasic.Config {
	return c.ConfigFunc()
}

func (c *ConfigStore) Parse(options zbasic.CommandLineOptions) {
	c.ParseFunc(options)
}

func (c *ConfigStore) GetByKey(key string) interface{} {
	return c.GetByKeyFunc(key)
}

func (c *ConfigStore) Validate(rules map[string]func(c *zbasic.Config) error) error {
	return c.ValidateFunc(rules)
}

func (c *ConfigStore) UnmarshalByKey(key string, model interface{}) error {
	return c.UnmarshalByKeyFunc(key, model)
}
