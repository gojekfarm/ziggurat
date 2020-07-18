package ziggurat

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
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
}

func ReadConfig() (*ZigguratConfig, error) {
	viper.SetConfigFile(defaultPath)
	viper.SetEnvPrefix("ZIGGURAT")
	var zigguratConfig ZigguratConfig
	if err := viper.ReadInConfig(); err != nil {
		if err, ok := err.(viper.ConfigFileNotFoundError); ok {
			errStr := fmt.Sprintf("error: config file not found %v", err)
			log.Fatal(errStr)
		} else {
			log.Printf("config error: %v", err)
			return nil, err
		}
	}

	err := viper.Unmarshal(&zigguratConfig)
	if err != nil {
		fmt.Printf("config error: %v\n", err)
	}
	return &zigguratConfig, nil
}
