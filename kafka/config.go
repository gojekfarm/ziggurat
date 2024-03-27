package kafka

import "github.com/confluentinc/confluent-kafka-go/v2/kafka"

type ConsumerConfig struct {
	BootstrapServers      string
	DebugLevel            string
	GroupID               string
	Topics                []string
	AutoCommitInterval    int
	ConsumerCount         int
	PollTimeout           int
	AutoOffsetReset       string
	PartitionAssignment   string
	MaxPollIntervalMS     int
	AllowAutoCreateTopics bool
}

func (c ConsumerConfig) toConfigMap() kafka.ConfigMap {

	kafkaConfMap := kafka.ConfigMap{
		"bootstrap.servers":        c.BootstrapServers,
		"group.id":                 c.GroupID,
		"enable.auto.commit":       true,
		"debug":                    "consumer",
		"go.logs.channel.enable":   true,
		"enable.auto.offset.store": false,
		"auto.commit.interval.ms":  5000,
		"auto.offset.reset":        "latest",
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

	if c.MaxPollIntervalMS > 0 {
		kafkaConfMap["max.poll.interval.ms"] = c.MaxPollIntervalMS
	}

	kafkaConfMap["allow.auto.create.topics"] = c.AllowAutoCreateTopics

	return kafkaConfMap
}
