package ziggurat

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

const defaultPollTimeout = 100 * time.Millisecond

func createConsumer(consumerConfig *kafka.ConfigMap, topics []string) *kafka.Consumer {
	consumer, err := kafka.NewConsumer(consumerConfig)
	if err != nil {
		log.Error().Err(err)
	}
	if subscribeErr := consumer.SubscribeTopics(topics, nil); subscribeErr != nil {
		log.Error().Err(subscribeErr)
	}
	return consumer
}

func startConsumer(routerCtx context.Context, handlerFunc HandlerFunc, consumer *kafka.Consumer, instanceID string, wg *sync.WaitGroup) {
	log.Info().Msgf("starting consumer with instance-id: %s", instanceID)
	go func(routerCtx context.Context, c *kafka.Consumer, instanceID string, waitGroup *sync.WaitGroup) {
		doneCh := routerCtx.Done()
		for {
			select {
			case <-doneCh:
				if err := consumer.Close(); err != nil {
					log.Error().Err(err)
				}
				wg.Done()
				return
			default:
				msg, err := c.ReadMessage(defaultPollTimeout)
				if err != nil && err.(kafka.Error).Code() != kafka.ErrTimedOut {
				} else if err != nil && err.(kafka.Error).Code() == kafka.ErrAllBrokersDown {
					log.Error().Err(err)
					return
				}
				if msg != nil {
					log.Info().Msgf("processing message for consumer %s", instanceID)
					messageEvent := NewMessageEvent(msg.Key, msg.Value, *msg.TopicPartition.Topic, int(msg.TopicPartition.Partition), instanceID)
					handlerFunc(messageEvent)
					fmt.Println(msg.TopicPartition.Offset)
					_, cmtErr := consumer.CommitMessage(msg)
					if cmtErr != nil {
						log.Error().Err(cmtErr)
					}
				}
			}
		}
	}(routerCtx, consumer, instanceID, wg)
}

func StartConsumers(routerCtx context.Context, consumerConfig *kafka.ConfigMap, topicEntity string, topics []string, instances int, handlerFunc HandlerFunc, wg *sync.WaitGroup) []*kafka.Consumer {
	consumers := make([]*kafka.Consumer, 0, instances)
	for i := 0; i < instances; i++ {
		consumer := createConsumer(consumerConfig, topics)
		consumers = append(consumers, consumer)
		groupID, _ := consumerConfig.Get("group.id", "")
		instanceID := fmt.Sprintf("%s-%s-%d", topicEntity, groupID, i)
		wg.Add(1)
		startConsumer(routerCtx, handlerFunc, consumer, instanceID, wg)
	}
	return consumers
}
