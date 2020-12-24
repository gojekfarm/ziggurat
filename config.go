package ziggurat

type KafkaStream struct {
	BootstrapServers string
	OriginTopics     string
	ConsumerGroupID  string
	ConsumerCount    int
}

type StreamRoutes map[string]KafkaStream
