package ziggurat

type KafkaStreamConfig struct {
	BootstrapServers string
	OriginTopics     string
	ConsumerGroupID  string
	ConsumerCount    int
}

type StreamRoutes map[string]KafkaStreamConfig
