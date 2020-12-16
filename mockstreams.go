package ziggurat

type MockKStreams struct {
	StartFunc func(a App) (chan struct{}, error)
}

func NewKafkaStreams() *MockKStreams {
	return &MockKStreams{StartFunc: func(a App) (chan struct{}, error) {
		return make(chan struct{}), nil
	}}
}

func (k MockKStreams) Start(a App) (chan struct{}, error) {
	return k.StartFunc(a)
}

func (k MockKStreams) Stop() {

}
