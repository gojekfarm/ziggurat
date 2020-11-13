package zig

import (
	"github.com/spf13/viper"
	"strings"
)

type StreamRouterConfig struct {
	InstanceCount    int    `mapstructure:"instance-count"`
	BootstrapServers string `mapstructure:"bootstrap-servers"`
	OriginTopics     string `mapstructure:"origin-topics"`
	GroupID          string `mapstructure:"group-id"`
}

type RetryConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Count   int  `mapstructure:"count"`
}

type HTTPServerConfig struct {
	Port string `mapstructure:"port"`
}

type Config struct {
	StreamRouter map[string]StreamRouterConfig `mapstructure:"stream-router"`
	LogLevel     string                        `mapstructure:"log-level"`
	ServiceName  string                        `mapstructure:"service-name"`
	Retry        RetryConfig                   `mapstructure:"retry"`
	HTTPServer   HTTPServerConfig              `mapstructure:"http-server"`
}

type ViperConfig struct {
	v            *viper.Viper
	parsedConfig *Config
}

func NewViperConfig() *ViperConfig {
	return &ViperConfig{
		v: viper.New(),
	}
}

func (vc *ViperConfig) Validate(rules map[string]func(c *Config) error) error {
	for _, validationFn := range rules {
		if err := validationFn(vc.parsedConfig); err != nil {
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

func (vc *ViperConfig) Parse(options CommandLineOptions) {
	vc.v.SetConfigFile(options.ConfigFilePath)
	vc.v.SetEnvPrefix("ziggurat")
	vc.v.SetConfigType("yaml")
	vc.v.AutomaticEnv()
	logFatal(vc.v.ReadInConfig(), "failed to read from config file", nil)
	replacer := strings.NewReplacer("-", "_", ".", "_")
	vc.v.SetEnvKeyReplacer(replacer)
	logFatal(vc.v.Unmarshal(&vc.parsedConfig), "failed to parse app config", nil)
}

func (vc *ViperConfig) Config() *Config {
	return vc.parsedConfig
}
