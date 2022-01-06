package kafka

import "github.com/confluentinc/confluent-kafka-go/kafka"

type ConsumerConfig struct {
	BootstrapServers   string
	DebugLevel         string
	GroupID            string
	Topics             string
	AutoCommitInterval int
	ConsumerCount      int
	PollTimeout        int
	RouteGroup         string
	AutoOffsetReset    string
}

func (c ConsumerConfig) toConfigMap() kafka.ConfigMap {
	autoCommitInterval := 15000
	debugLevel := "consumer"
	autoOffsetReset := "earliest"

	if c.AutoCommitInterval > 0 {
		autoCommitInterval = c.AutoCommitInterval
	}
	if c.DebugLevel != "" {
		debugLevel = c.DebugLevel
	}

	if c.AutoOffsetReset != "" {
		autoOffsetReset = c.AutoOffsetReset
	}

	kafkaConfMap := kafka.ConfigMap{
		"bootstrap.servers":        c.BootstrapServers,
		"group.id":                 c.GroupID,
		"auto.offset.reset":        autoOffsetReset,
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  autoCommitInterval,
		"debug":                    debugLevel,
		"go.logs.channel.enable":   true,
		"enable.auto.offset.store": false,
	}
	return kafkaConfMap
}

type StreamConfig = []ConsumerConfig
