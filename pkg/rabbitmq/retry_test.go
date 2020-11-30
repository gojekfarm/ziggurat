package rabbitmq

import (
	"context"
	"errors"
	"github.com/gojekfarm/ziggurat-go/pkg/mock"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zbasic"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/publisher"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
	"os"
	"reflect"
	"testing"
	"time"
)

const entityName = "foo"
const retryCountValue = 3

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	os.Exit(m.Run())
}

func TestRabbitMQRetry_StartWithDialerError(t *testing.T) {
	cs := mock.NewConfigStore()
	app := mock.NewApp()
	rmq := NewRabbitMQRetry(cs)
	dialerError := errors.New("dialer error")
	oldCreateDialer := createDialer
	defer func() {
		createDialer = oldCreateDialer
	}()
	createDialer = func(ctx context.Context, hosts []string) (*amqpextra.Dialer, error) {
		return nil, dialerError
	}
	err := rmq.Start(app)
	if err != dialerError {
		t.Errorf("expected error to be %v got %v", dialerError, err)
	}
}

func TestRabbitMQRetry_StartSuccess(t *testing.T) {
	cs := mock.NewConfigStore()
	expectedServiceName := "baz"
	expectedEntities := []string{entityName}
	app := mock.NewApp()
	rmq := NewRabbitMQRetry(cs)
	cs.ConfigFunc = func() *zbasic.Config {
		return &zbasic.Config{
			ServiceName: "baz",
			Retry: zbasic.RetryConfig{
				Enabled: true,
				Count:   retryCountValue,
			},
		}
	}
	app.ConfigStoreFunc = func() z.ConfigStore {
		return cs
	}
	app.RoutesFunc = func() []string {
		return []string{"foo"}
	}
	createDialer = func(ctx context.Context, hosts []string) (*amqpextra.Dialer, error) {
		return &amqpextra.Dialer{}, nil
	}
	setupConsumers = func(app z.App, dialer *amqpextra.Dialer) error {
		return nil
	}
	getConnectionFromDialer = func(ctx context.Context, d *amqpextra.Dialer, timeout time.Duration) (*amqp.Connection, error) {
		return &amqp.Connection{}, nil
	}
	withChannel = func(connection *amqp.Connection, cb func(c *amqp.Channel) error) error {
		cb(&amqp.Channel{})
		return nil
	}
	createAndBindQueues = func(c *amqp.Channel, topicEntities []string, serviceName string) {
		if serviceName != expectedServiceName {
			t.Errorf("expected servicename %s got %s", expectedServiceName, serviceName)
		}
		if !reflect.DeepEqual(topicEntities, expectedEntities) {
			t.Errorf("expected entities to be %v got %v", expectedEntities, topicEntities)
		}
	}
	err := rmq.Start(app)
	if err != nil {
		t.Errorf("expected error to nil")
	}
}

func TestRabbitMQRetry_RetryDelayQueue(t *testing.T) {
	cs := mock.NewConfigStore()
	cs.ConfigFunc = func() *zbasic.Config {
		return &zbasic.Config{
			ServiceName: "baz",
			Retry: zbasic.RetryConfig{
				Enabled: true,
				Count:   3,
			},
		}
	}

	message := zbasic.NewMessageEvent(nil, nil, "", "foo", "", time.Time{})
	createPublisher = func(ctx context.Context, d *amqpextra.Dialer) (*publisher.Publisher, error) {
		ch := make(<-chan *publisher.Connection)
		return publisher.New(ch)
	}
	publishMessage = func(exchangeName string, p *publisher.Publisher, payload zbasic.MessageEvent, expirationInMS string) error {
		expectedExchangeName := "foo_baz_delay_exchange"
		if exchangeName != expectedExchangeName {
			t.Errorf("expected exchange name %s got %s", expectedExchangeName, exchangeName)
		}
		return nil
	}
	app := mock.NewApp()
	app.ConfigStoreFunc = func() z.ConfigStore {
		return cs
	}
	rmq := NewRabbitMQRetry(cs)

	rmq.Retry(app, message)
}

func TestRabbitMQRetry_RetryDLQueue(t *testing.T) {
	cs := mock.NewConfigStore()
	cs.ConfigFunc = func() *zbasic.Config {
		return &zbasic.Config{
			ServiceName: "baz",
			Retry: zbasic.RetryConfig{
				Enabled: true,
				Count:   3,
			},
		}
	}
	message := zbasic.NewMessageEvent(nil, nil, "", "foo", "", time.Time{})
	message.SetMessageAttribute(retryCount, retryCountValue)
	createPublisher = func(ctx context.Context, d *amqpextra.Dialer) (*publisher.Publisher, error) {
		ch := make(<-chan *publisher.Connection)
		return publisher.New(ch)
	}
	publishMessage = func(exchangeName string, p *publisher.Publisher, payload zbasic.MessageEvent, expirationInMS string) error {
		expectedExchangeName := "foo_baz_dead_letter_exchange"
		if exchangeName != expectedExchangeName {
			t.Errorf("expected exchange name %s got %s", expectedExchangeName, exchangeName)
		}
		return nil
	}
	app := mock.NewApp()
	app.ConfigStoreFunc = func() z.ConfigStore {
		return cs
	}
	rmq := NewRabbitMQRetry(cs)

	rmq.Retry(app, message)
}
