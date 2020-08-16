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

func (config *Config) Validate() {
	if len(config.StreamRouters) == 0 {
		log.Fatal().Str("stream-router-count", "0").Msg("no stream routers specified, exiting app...")
	}
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
