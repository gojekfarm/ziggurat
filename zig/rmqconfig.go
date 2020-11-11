package zig

import "errors"

type RabbitMQConfig struct {
	host                 string
	delayQueueExpiration string
}

func parseRabbitMQConfig(config ConfigReader) *RabbitMQConfig {
	rawConfig := config.GetByKey("rabbitmq")
	if sanitizedConfig, ok := rawConfig.(map[string]interface{}); !ok {
		logError(errors.New("errors parsing rabbitmq config"), "rmq config: parse error", nil)
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
