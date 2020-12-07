package retry

import (
	"github.com/streadway/amqp"
	"reflect"
	"testing"
)

func TestConstructQueueName(t *testing.T) {
	expectedQueueName := "foo_bar_instant_queue"
	queueName := constructQueueName("bar", "foo", QueueTypeInstant)
	if queueName != expectedQueueName {
		t.Errorf("expected queue name to be %s , got %s", expectedQueueName, queueName)
	}
}

func TestConstructExchangeName(t *testing.T) {
	expectedExchangeName := "foo_bar_instant_exchange"
	exchangeName := constructExchangeName("bar", "foo", QueueTypeInstant)
	if exchangeName != expectedExchangeName {
		t.Errorf("expected queue name to be %s , got %s", expectedExchangeName, exchangeName)
	}
}

func TestDeclareExchanges(t *testing.T) {
	routes := []string{"foo", "bar"}
	callCount := 0
	serviceName := "baz"
	queueTypeCount := 3
	expectedCallCount := len(routes) * queueTypeCount
	declareExchange = func(c *amqp.Channel, exchangeName string) {
		callCount++
	}
	declareExchanges(&amqp.Channel{}, routes, serviceName)
	if callCount != expectedCallCount {
		t.Errorf("expected call count to be %d got %d", expectedCallCount, callCount)
	}
}

func TestDelayQueueCreateAndBind(t *testing.T) {
	serviceName, routes := "baz", []string{"foo"}
	expectedArgs := amqp.Table{"x-dead-letter-exchange": constructExchangeName(serviceName, routes[0], QueueTypeInstant)}
	queueDeclare = func(c *amqp.Channel, queueName string, args amqp.Table) (amqp.Queue, error) {
		expectedQueueName := "foo_baz_" + QueueTypeDelay + "_queue"
		if queueName != expectedQueueName {
			t.Errorf("expected queue name %s got %s", expectedQueueName, queueName)
		}
		if !reflect.DeepEqual(args, expectedArgs) {
			t.Errorf("expected args %v, got %v", expectedArgs, args)
		}
		return amqp.Queue{}, nil
	}

	queueBind = func(c *amqp.Channel, queueName string, exchangeName string, args amqp.Table) error {
		expectedQueueName := "foo_baz_" + QueueTypeDelay + "_queue"
		expectedExchangeName := "foo_baz_" + QueueTypeDelay + "_exchange"
		if queueName != expectedQueueName {
			t.Errorf("expected queue name %s got %s", expectedQueueName, queueName)
		}
		if exchangeName != expectedExchangeName {
			t.Errorf("expected exchange name %s got %s", expectedExchangeName, exchangeName)
		}
		if !reflect.DeepEqual(args, expectedArgs) {
			t.Errorf("expected args %v, got %v", expectedArgs, args)
		}
		return nil
	}
	createDelayQueues(&amqp.Channel{}, routes, serviceName)
}
