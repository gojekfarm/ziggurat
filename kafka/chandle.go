package kafka

import (
	"errors"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func ConsumerHandle(group *ConsumerGroup) (*kafka.Consumer, error) {
	if group.c == nil {
		return nil, errors.New("consumer handle not initialized")
	}
	kc, ok := group.c.(*kafka.Consumer)
	if !ok {
		return nil, errors.New("could not assert to *kafka.Consumer")
	}
	return kc, nil
}
