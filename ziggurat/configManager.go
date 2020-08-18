package ziggurat

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const defaultPath = "./config/config.yaml"

type StreamRouterConfig struct {
	InstanceCount    int    `mapstructure:"instance-count"`
	BootstrapServers string `mapstructure:"bootstrap-servers"`
	OriginTopics     string `mapstructure:"origin-topics"`
	GroupID          string `mapstructure:"group-id"`
	TopicEntity      string `mapstructure:"topic-entity"`
}

type Config struct {
	StreamRouters []StreamRouterConfig `mapstructure:"stream-router"`
	LogLevel      string               `mapstructure:"log-level"`
	ServiceName   string               `mapstructure:"service-name"`
}

var zigguratConfig Config

var configValidationRuleMapping = map[string]func(c *Config) error{
	"streamRouteValidation": func(c *Config) error {
		if len(c.StreamRouters) == 0 {
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

func (config *Config) Validate() error {
	for _, validationFn := range configValidationRuleMapping {
		if err:=validationFn(config);err!=nil{
			return err
		}
	}
	return nil
}

func parseConfig() {
	viper.SetConfigFile(defaultPath)
	viper.SetEnvPrefix("ZIGGURAT")
	if err := viper.ReadInConfig(); err != nil {
		if err, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal().Err(err)
		}
	}

	if err := viper.Unmarshal(&zigguratConfig); err != nil {
		log.Fatal().Err(err)
	}
}

func GetConfig() Config {
	return zigguratConfig
}
