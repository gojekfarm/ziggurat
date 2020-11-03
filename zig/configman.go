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

var zigguratConfig Config

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

func (config *Config) validate() error {
	for _, validationFn := range configValidationRuleMapping {
		if err := validationFn(config); err != nil {
			return err
		}
	}
	return nil
}

func (config *Config) GetByKey(key string) interface{} {
	return viper.Get(key)
}

func parseConfig(options CommandLineOptions) {
	viper.SetConfigFile(options.ConfigFilePath)
	viper.SetEnvPrefix("ziggurat")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("[ZIG APP]")
	}

	replacer := strings.NewReplacer("-", "_", ".", "_")
	viper.SetEnvKeyReplacer(replacer)
	if err := viper.Unmarshal(&zigguratConfig); err != nil {
		log.Fatal().Err(err).Msg("[ZIG APP] config parse error")
	}
}

func getConfig() Config {
	return zigguratConfig
}
