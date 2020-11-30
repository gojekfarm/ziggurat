package kstream

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat-go/pkg/mock"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zb"
	"sync"
	"testing"
)

func TestKafkaStreams_Start(t *testing.T) {
	a := mock.NewApp()
	a.ConfigStoreFunc = func() z.ConfigStore {
		c := mock.NewConfigStore()
		c.ConfigFunc = func() *zb.Config {
			return &zb.Config{
				StreamRouter: map[string]zb.StreamRouterConfig{
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

	StartConsumers = func(app z.App, consumerConfig *kafka.ConfigMap, topicEntity string, topics []string, instances int, h z.MessageHandler, wg *sync.WaitGroup) []*kafka.Consumer {
		wg.Add(1)
		wg.Done()
		return []*kafka.Consumer{}
	}
	kstreams := NewKafkaStreams()
	done, _ := kstreams.Start(a)
	if len(kstreams.routeConsumerMap) < len(a.Routes()) {
		t.Errorf("expected count %d but got %d", len(kstreams.routeConsumerMap), len(a.Routes()))
	}
	<-done

}
