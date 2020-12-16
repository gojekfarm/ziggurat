package ziggurat

type StreamRouterConfig struct {
	InstanceCount    int    `mapstructure:"instance-count"`
	BootstrapServers string `mapstructure:"bootstrap-servers"`
	OriginTopics     string `mapstructure:"origin-topics"`
	GroupID          string `mapstructure:"group-id"`
}

type Stream struct {
	InstanceCount    int
	BootstrapServers string
	OriginTopics     string
	GroupID          string
}

type Routes map[string]Stream
