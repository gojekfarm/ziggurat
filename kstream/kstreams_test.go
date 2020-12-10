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
	a := mock.NewApp()
	a.ConfigStoreFunc = func() ztype.ConfigStore {
		c := mock.NewConfigStore()
		c.ConfigFunc = func() *zbase.Config {
			return &zbase.Config{
				StreamRouter: map[string]zbase.StreamRouterConfig{
					"foo": {
						InstanceCount:    0,
						BootstrapServers: "",
						OriginTopics:     "",
						GroupID:          "",
					},
					"bar": {
						InstanceCount:    0,
						BootstrapServers: "",
						OriginTopics:     "",
						GroupID:          "",
					},
				},
			}
		}
		return c
	}
	a.RoutesFunc = func() []string {
		return []string{"foo", "bar"}
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
