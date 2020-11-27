package kstream

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zbasic"
	"github.com/gojekfarm/ziggurat-go/pkg/zlogger"
	"github.com/gojekfarm/ziggurat-go/pkg/zmw"
	"sync"
	"time"
)

const defaultPollTimeout = 100 * time.Millisecond
const brokerRetryTimeout = 2 * time.Second

var startConsumer = func(ctx context.Context, app z.App, h z.MessageHandler, consumer *kafka.Consumer, topicEntity string, instanceID string, wg *sync.WaitGroup) {
	zlogger.LogInfo("consumer: starting consumer", map[string]interface{}{"consumer-instance-id": instanceID})
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
					zlogger.LogError(err, "consumer: retrying broker...", nil)
					time.Sleep(brokerRetryTimeout)
					continue
				}
				if msg != nil {
					messageEvent := zbasic.NewMessageEvent(msg.Key, msg.Value, *msg.TopicPartition.Topic, topicEntity, msg.TimestampType.String(), msg.Timestamp)
					h.HandleMessage(messageEvent, app)
					zlogger.LogError(storeOffsets(consumer, msg.TopicPartition), "ziggurat consumer", nil)
				}
			}
		}
	}(ctx, consumer, instanceID, wg)
}

var StartConsumers = func(routerCtx context.Context, app z.App, consumerConfig *kafka.ConfigMap, topicEntity string, topics []string, instances int, h z.MessageHandler, wg *sync.WaitGroup) []*kafka.Consumer {
	consumers := make([]*kafka.Consumer, 0, instances)
	for i := 0; i < instances; i++ {
		consumer := createConsumer(consumerConfig, topics)
		consumers = append(consumers, consumer)
		groupID, _ := consumerConfig.Get("group.id", "")
		instanceID := fmt.Sprintf("%s_%s_%d", topicEntity, groupID, i)
		wg.Add(1)
		startConsumer(routerCtx, app, zmw.DefaultTerminalMW(h), consumer, topicEntity, instanceID, wg)
	}
	return consumers
}
