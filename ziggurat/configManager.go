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

type ZigguratConfig struct {
	StreamRouters []StreamRouterConfig `mapstructure:"stream-router"`
	LogLevel      string               `mapstructure:"log-level"`
}

var zigguratConfig ZigguratConfig

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

func GetConfig() ZigguratConfig {
	return zigguratConfig
}
