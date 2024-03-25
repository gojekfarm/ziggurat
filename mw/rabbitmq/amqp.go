package rabbitmq

import (
	"github.com/makasim/amqpextra/publisher"
	"github.com/stretchr/testify/mock"
)

type amqpPublisher interface {
	Publish(message publisher.Message) error
}

type mockAMQPPublisher struct {
	mock.Mock
}

func (m *mockAMQPPublisher) Publish(message publisher.Message) error {
	return m.Called(message).Error(0)
}
