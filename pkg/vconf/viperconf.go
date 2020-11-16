package vconf

import (
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	"github.com/spf13/viper"
	"strings"
)

type ViperConfig struct {
	v            *viper.Viper
	ParsedConfig *basic.Config
}

func NewViperConfig() *ViperConfig {
	return &ViperConfig{
		v: viper.New(),
	}
}

func (vc *ViperConfig) Validate(rules map[string]func(c *basic.Config) error) error {
	for _, validationFn := range rules {
		if err := validationFn(vc.ParsedConfig); err != nil {
			return err
		}
	}
	return nil
}

func (vc *ViperConfig) GetByKey(key string) interface{} {
	return vc.v.Get(key)
}

func (vc *ViperConfig) UnmarshalByKey(key string, model interface{}) error {
	return vc.v.UnmarshalKey(key, model)
}

func (vc *ViperConfig) Parse(options basic.CommandLineOptions) {
	vc.v.SetConfigFile(options.ConfigFilePath)
	vc.v.SetEnvPrefix("ziggurat")
	vc.v.SetConfigType("yaml")
	vc.v.AutomaticEnv()
	logger.LogFatal(vc.v.ReadInConfig(), "failed to read from config file", nil)
	replacer := strings.NewReplacer("-", "_", ".", "_")
	vc.v.SetEnvKeyReplacer(replacer)
	logger.LogFatal(vc.v.Unmarshal(&vc.ParsedConfig), "failed to parse app config", nil)
}

func (vc *ViperConfig) Config() *basic.Config {
	return vc.ParsedConfig
}
