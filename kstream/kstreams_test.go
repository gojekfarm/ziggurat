package kstream

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat/mock"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/ztype"
	"sync"
	"testing"
)

func TestKafkaStreams_Start(t *testing.T) {
	a := mock.NewZig()
	a.RoutesFunc = func() zbase.Routes {
		return zbase.Routes{"foo": {}}
	}
	oldStartConsumers := StartConsumers
	defer func() {
		StartConsumers = oldStartConsumers
	}()

	StartConsumers = func(app ztype.App, consumerConfig *kafka.ConfigMap, topicEntity string, topics []string, instances int, h ztype.MessageHandler, wg *sync.WaitGroup) []*kafka.Consumer {
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
