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

func storeOffsets(consumer *kafka.Consumer, partition kafka.TopicPartition) error {
	if partition.Error != nil {
		return ErrOffsetCommit
	}
	offsets := []kafka.TopicPartition{partition}
	offsets[0].Offset++
	if _, err := consumer.StoreOffsets(offsets); err != nil {
		return err
	}
	return nil
}

func startConsumer(routerCtx context.Context, config Config, handlerFunc HandlerFunc, consumer *kafka.Consumer, topicEntity string, instanceID string, retrier MessageRetrier, wg *sync.WaitGroup) {
	log.Info().Msgf("starting consumer with instance-id: %s", instanceID)

	go func(routerCtx context.Context, c *kafka.Consumer, instanceID string, waitGroup *sync.WaitGroup) {
		doneCh := routerCtx.Done()
		for {
			select {
			case <-doneCh:
				if err := consumer.Close(); err != nil {
					log.Error().Err(err).Msg("error closing consumer")
				}
				log.Info().Str("consumer-instance-id", instanceID).Msg("stopping consumer...")
				wg.Done()
				return
			default:
				msg, err := c.ReadMessage(defaultPollTimeout)
				if err != nil && err.(kafka.Error).Code() != kafka.ErrTimedOut {
				} else if err != nil && err.(kafka.Error).Code() == kafka.ErrAllBrokersDown {
					log.Error().Err(err).Msg("terminating application, all brokers down")
					return
				}
				if msg != nil {
					log.Info().Msgf("processing message for consumer %s", instanceID)
					messageEvent := MessageEvent{
						MessageKey:        nil,
						MessageValue:      nil,
						MessageValueBytes: msg.Key,
						MessageKeyBytes:   msg.Value,
						Topic:             *msg.TopicPartition.Topic,
						TopicEntity:       topicEntity,
						Attributes:        make(map[string]interface{}),
					}
					MessageHandler(config, handlerFunc, retrier)(messageEvent)
					if commitErr := storeOffsets(consumer, msg.TopicPartition); commitErr != nil {
						log.Error().Err(commitErr).Msg("offset commit error")
					}

				}
			}
		}
	}(routerCtx, consumer, instanceID, wg)
}

func StartConsumers(routerCtx context.Context, config Config, consumerConfig *kafka.ConfigMap, topicEntity string, topics []string, instances int, handlerFunc HandlerFunc, retrier MessageRetrier, wg *sync.WaitGroup) []*kafka.Consumer {
	consumers := make([]*kafka.Consumer, 0, instances)
	for i := 0; i < instances; i++ {
		consumer := createConsumer(consumerConfig, topics)
		consumers = append(consumers, consumer)
		groupID, _ := consumerConfig.Get("group.id", "")
		instanceID := fmt.Sprintf("%s-%s-%d", topicEntity, groupID, i)
		wg.Add(1)
		startConsumer(routerCtx, config, handlerFunc, consumer, topicEntity, instanceID, retrier, wg)
	}
	return consumers
}
