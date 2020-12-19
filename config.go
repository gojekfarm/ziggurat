package ziggurat

type Stream struct {
	InstanceCount    int
	BootstrapServers string
	OriginTopics     string
	GroupID          string
}

type Routes map[string]Stream
