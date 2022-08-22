package kafka

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Consumer interface {
	Consume(ctx context.Context) (*kafka.Message, error)
	Close() error
}
