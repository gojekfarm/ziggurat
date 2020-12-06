package mock

import (
	"github.com/gojekfarm/ziggurat-go/pkg/z"
)

type KafkaStreams struct {
	StartFunc func(a z.App) (chan struct{}, error)
}

func NewKafkaStreams() *KafkaStreams {
	return &KafkaStreams{StartFunc: func(a z.App) (chan struct{}, error) {
		return make(chan struct{}), nil
	}}
}

func (k KafkaStreams) Start(a z.App) (chan struct{}, error) {
	return k.StartFunc(a)
}

func (k KafkaStreams) Stop() {

}
