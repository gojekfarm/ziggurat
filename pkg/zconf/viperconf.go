package zconf

import (
	"github.com/gojekfarm/ziggurat-go/pkg/zbasic"
	"github.com/spf13/viper"
	"strings"
)

const configFormat = "yaml"
const envPrefix = "ziggurat"

type ViperConfig struct {
	v            *viper.Viper
	ParsedConfig *zbasic.Config
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

func (vc *ViperConfig) Parse(options zbasic.CommandLineOptions) error {
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

func (vc *ViperConfig) Config() *zbasic.Config {
	return vc.ParsedConfig
}
