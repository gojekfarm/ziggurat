package zig

import (
	"errors"
	"github.com/rs/zerolog/log"
)

type RabbitMQConfig struct {
	host                 string
	delayQueueExpiration string
}

func parseRabbitMQConfig(config *Config) *RabbitMQConfig {
	rawConfig := config.GetByKey("rabbitmq")
	if sanitizedConfig, ok := rawConfig.(map[string]interface{}); !ok {
		log.Error().Err(errors.New("error parsing rabbitmq appconf")).Msg("")
		return &RabbitMQConfig{
			host:                 "amqp://user:guest@localhost:5672/",
			delayQueueExpiration: "2000",
		}
	} else {
		return &RabbitMQConfig{
			host:                 sanitizedConfig["host"].(string),
			delayQueueExpiration: sanitizedConfig["delay-queue-expiration"].(string),
		}
	}
}
