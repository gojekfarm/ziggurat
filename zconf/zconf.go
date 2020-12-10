package zconf

import (
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/spf13/viper"
	"strings"
)

const configFormat = "yaml"
const envPrefix = "ziggurat"

type ViperConfig struct {
	v            *viper.Viper
	ParsedConfig *zbase.Config
}

func NewViperConfig() *ViperConfig {
	return &ViperConfig{
		v: viper.New(),
	}
}

func (vc *ViperConfig) GetByKey(key string) interface{} {
	return vc.v.Get(key)
}

func (vc *ViperConfig) UnmarshalByKey(key string, model interface{}) error {
	return vc.v.UnmarshalKey(key, model)
}

func (vc *ViperConfig) Parse(options zbase.CommandLineOptions) error {
	vc.v.SetConfigFile(options.ConfigFilePath)
	vc.v.SetEnvPrefix(envPrefix)
	vc.v.SetConfigType(configFormat)
	vc.v.AutomaticEnv()
	if err := vc.v.ReadInConfig(); err != nil {
		return err
	}
	replacer := strings.NewReplacer("-", "_", ".", "_")
	vc.v.SetEnvKeyReplacer(replacer)
	if err := vc.v.Unmarshal(&vc.ParsedConfig); err != nil {
		return err
	}
	return nil
}

func (vc *ViperConfig) Config() *zbase.Config {
	return vc.ParsedConfig
}
