package ziggurat

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"sync"
	"testing"
)

func TestKafkaStreams_Start(t *testing.T) {
	a := NewZig()
	a.RoutesFunc = func() Routes {
		return Routes{"foo": Stream{}}
	}
	oldStartConsumers := StartConsumers
	defer func() {
		StartConsumers = oldStartConsumers
	}()

	StartConsumers = func(app AppContext, consumerConfig *kafka.ConfigMap, topicEntity string, topics []string, instances int, h MessageHandler, wg *sync.WaitGroup) []*kafka.Consumer {
		wg.Add(1)
		wg.Done()
		return []*kafka.Consumer{}
	}
	kstreams := New()
	done, _ := kstreams.Start(a)
	if len(kstreams.routeConsumerMap) < len(a.Routes()) {
		t.Errorf("expected count %d but got %d", len(kstreams.routeConsumerMap), len(a.Routes()))
	}
	<-done

}
