package kafka

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Consumer interface {
	Consume(ctx context.Context) (*kafka.Message, error)
	Close() error
}
