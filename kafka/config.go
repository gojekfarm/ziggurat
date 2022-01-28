package kafka

import "github.com/confluentinc/confluent-kafka-go/kafka"

type ConsumerConfig struct {
	BootstrapServers    string
	DebugLevel          string
	GroupID             string
	Topics              string
	AutoCommitInterval  int
	ConsumerCount       int
	PollTimeout         int
	RouteGroup          string
	AutoOffsetReset     string
	PartitionAssignment string
}

func (c ConsumerConfig) toConfigMap() kafka.ConfigMap {
	autoCommitInterval := 15000
	debugLevel := "consumer"
	autoOffsetReset := "earliest"
	partitionAssignment := "range,roundrobin"

	if c.AutoCommitInterval > 0 {
		autoCommitInterval = c.AutoCommitInterval
	}
	if c.DebugLevel != "" {
		debugLevel = c.DebugLevel
	}

	if c.AutoOffsetReset != "" {
		autoOffsetReset = c.AutoOffsetReset
	}

	if c.PartitionAssignment != "" {
		partitionAssignment = c.PartitionAssignment
	}

	kafkaConfMap := kafka.ConfigMap{
		"bootstrap.servers":             c.BootstrapServers,
		"group.id":                      c.GroupID,
		"auto.offset.reset":             autoOffsetReset,
		"enable.auto.commit":            true,
		"auto.commit.interval.ms":       autoCommitInterval,
		"debug":                         debugLevel,
		"go.logs.channel.enable":        true,
		"enable.auto.offset.store":      false,
		"partition.assignment.strategy": partitionAssignment,
	}
	return kafkaConfMap
}

type StreamConfig = []ConsumerConfig
