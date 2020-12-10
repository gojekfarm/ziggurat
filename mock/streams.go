package mock

import (
	"github.com/gojekfarm/ziggurat/ztype"
)

type KafkaStreams struct {
	StartFunc func(a ztype.App) (chan struct{}, error)
}

func NewKafkaStreams() *KafkaStreams {
	return &KafkaStreams{StartFunc: func(a ztype.App) (chan struct{}, error) {
		return make(chan struct{}), nil
	}}
}

func (k KafkaStreams) Start(a ztype.App) (chan struct{}, error) {
	return k.StartFunc(a)
}

func (k KafkaStreams) Stop() {

}
