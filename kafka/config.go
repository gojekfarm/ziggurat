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

	kafkaConfMap := kafka.ConfigMap{
		"bootstrap.servers":        c.BootstrapServers,
		"group.id":                 c.GroupID,
		"enable.auto.commit":       true,
		"debug":                    "consumer",
		"go.logs.channel.enable":   true,
		"enable.auto.offset.store": false,
		"auto.commit.interval.ms":  15000,
		"auto.offset.reset":        "earliest",
	}
	if c.AutoOffsetReset != "" {
		kafkaConfMap["auto.offset.reset"] = c.AutoOffsetReset
	}

	if c.PartitionAssignment != "" {
		kafkaConfMap["partition.assignment.strategy"] = c.PartitionAssignment
	}

	if c.DebugLevel != "" {
		kafkaConfMap["debug"] = c.DebugLevel
	}

	if c.AutoCommitInterval > 0 {
		kafkaConfMap["auto.commit.interval.ms"] = c.AutoCommitInterval
	}

	return kafkaConfMap
}

type StreamConfig = []ConsumerConfig
