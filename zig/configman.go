package zig

import (
	"github.com/rs/zerolog/log"
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

var configValidationRuleMapping = map[string]func(c *Config) error{
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

func (vc *ViperConfig) Validate() error {
	for _, validationFn := range configValidationRuleMapping {
		if err := validationFn(vc.parsedConfig); err != nil {
			return err
		}
	}
	return nil
}

func (vc *ViperConfig) GetByKey(key string) interface{} {
	return vc.v.Get(key)
}

func (vc *ViperConfig) Parse(options CommandLineOptions) {
	vc.v.SetConfigFile(options.ConfigFilePath)
	vc.v.SetEnvPrefix("ziggurat")
	vc.v.SetConfigType("yaml")
	vc.v.AutomaticEnv()
	if err := vc.v.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("")
	}

	replacer := strings.NewReplacer("-", "_", ".", "_")
	vc.v.SetEnvKeyReplacer(replacer)
	if err := vc.v.Unmarshal(&vc.parsedConfig); err != nil {
		log.Fatal().Err(err).Msg(" appconf parse error")
	}
}

func (vc *ViperConfig) Config() *Config {
	return vc.parsedConfig
}
