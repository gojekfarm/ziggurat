package cons

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/handler"
	"github.com/gojekfarm/ziggurat-go/pkg/logger"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"sync"
	"time"
)

var consumerLogContext = map[string]interface{}{"component": "consumer"}

const defaultPollTimeout = 100 * time.Millisecond
const brokerRetryTimeout = 2 * time.Second

func startConsumer(ctx context.Context, app z.App, handlerFunc z.HandlerFunc, consumer *kafka.Consumer, topicEntity string, instanceID string, wg *sync.WaitGroup) {
	logger.LogInfo("ziggurat consumer: starting consumer", map[string]interface{}{"consumer-instance-id": instanceID})
	go func(routerCtx context.Context, c *kafka.Consumer, instanceID string, waitGroup *sync.WaitGroup) {
		doneCh := routerCtx.Done()
		for {
			select {
			case <-doneCh:
				wg.Done()
				return
			default:
				msg, err := readMessage(c, defaultPollTimeout)
				if err != nil && err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				} else if err != nil && err.(kafka.Error).Code() == kafka.ErrAllBrokersDown {
					logger.LogError(err, "consumer: retrying broker...", nil)
					time.Sleep(brokerRetryTimeout)
					continue
				}
				if msg != nil {
					messageEvent := basic.NewMessageEvent(msg.Key, msg.Value, *msg.TopicPartition.Topic, topicEntity, msg.TimestampType.String(), msg.Timestamp)
					handler.MessageHandler(app, handlerFunc)(messageEvent)
					logger.LogError(storeOffsets(consumer, msg.TopicPartition), "ziggurat consumer", nil)
				}
			}
		}
	}(ctx, consumer, instanceID, wg)
}

var StartConsumers = func(routerCtx context.Context, app z.App, consumerConfig *kafka.ConfigMap, topicEntity string, topics []string, instances int, handlerFunc z.HandlerFunc, wg *sync.WaitGroup) []*kafka.Consumer {
	consumers := make([]*kafka.Consumer, 0, instances)
	for i := 0; i < instances; i++ {
		consumer := createConsumer(consumerConfig, topics)
		consumers = append(consumers, consumer)
		groupID, _ := consumerConfig.Get("group.id", "")
		instanceID := fmt.Sprintf("%s-%s-%d", topicEntity, groupID, i)
		wg.Add(1)
		startConsumer(routerCtx, app, handlerFunc, consumer, topicEntity, instanceID, wg)
	}
	return consumers
}
