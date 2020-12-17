package ziggurat

type Stream struct {
	InstanceCount    int
	BootstrapServers string
	OriginTopics     string
	GroupID          string
}

func (s Stream) ThreadCount() int {
	return s.InstanceCount
}

func (s Stream) Servers() string {
	return s.BootstrapServers
}

func (s Stream) ConsumerGroupID() string {
	return s.GroupID
}

func (s Stream) Topics() string {
	return s.OriginTopics
}
