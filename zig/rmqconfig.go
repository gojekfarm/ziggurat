package zig

import "fmt"

type RabbitMQConfig struct {
	Host                 string `mapstructure:"host"`
	DelayQueueExpiration string `mapstructure:"delay-queue-expiration"`
}

func parseRabbitMQConfig(config ConfigReader) *RabbitMQConfig {
	rmqcfg := &RabbitMQConfig{}
	if err := config.UnmarshalByKey("rabbitmq", rmqcfg); err != nil {
		logError(err, "rmq config unmarshall error", nil)
		return &RabbitMQConfig{
			Host:                 "amqp://user:bitnami@localhost:5672/",
			DelayQueueExpiration: "2000",
		}
	}
	fmt.Printf("RMQ_CFG :=> %+v", rmqcfg)
	return rmqcfg
}
