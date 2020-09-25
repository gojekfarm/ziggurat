package ziggurat

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"sync"
	"time"
)

const DefaultPollTimeout = 100 * time.Millisecond

func createConsumer(consumerConfig *kafka.ConfigMap, topics []string) *kafka.Consumer {
	consumer, err := kafka.NewConsumer(consumerConfig)
	if err != nil {
		ConsumerLogger.Error().Err(err)
	}
	if subscribeErr := consumer.SubscribeTopics(topics, nil); subscribeErr != nil {
		ConsumerLogger.Error().Err(subscribeErr)
	}
	return consumer
}

func storeOffsets(consumer *kafka.Consumer, partition kafka.TopicPartition) error { // atleast once delivery
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

func startConsumer(routerCtx context.Context, applicationContext ApplicationContext, handlerFunc HandlerFunc, consumer *kafka.Consumer, topicEntity string, instanceID string, wg *sync.WaitGroup) {
	ConsumerLogger.Info().Str("consumer-instance-id", instanceID).Msg("starting consumer")

	go func(routerCtx context.Context, c *kafka.Consumer, instanceID string, waitGroup *sync.WaitGroup) {
		doneCh := routerCtx.Done()
		for {
			select {
			case <-doneCh:
				if err := consumer.Close(); err != nil {
					ConsumerLogger.Error().Err(err).Msg("error closing consumer")
				}
				ConsumerLogger.Info().Str("consumer-instance-id", instanceID).Msg("stopping consumer")
				wg.Done()
				return
			default:
				msg, err := c.ReadMessage(DefaultPollTimeout)
				if err != nil && err.(kafka.Error).Code() != kafka.ErrTimedOut {
				} else if err != nil && err.(kafka.Error).Code() == kafka.ErrAllBrokersDown {
					ConsumerLogger.Error().Err(err).Msg("terminating application, all brokers down")
					return
				}
				if msg != nil {
					ConsumerLogger.Info().Msgf("processing message for consumer %s", instanceID)
					messageEvent := MessageEvent{
						MessageKey:        nil,
						MessageValue:      nil,
						MessageValueBytes: msg.Key,
						MessageKeyBytes:   msg.Value,
						Topic:             *msg.TopicPartition.Topic,
						TopicEntity:       topicEntity,
						TimestampType:     msg.TimestampType.String(),
						KafkaTimestamp:    msg.Timestamp,
						Attributes:        make(map[string]interface{}),
					}
					MessageHandler(applicationContext, handlerFunc, applicationContext.Retrier)(messageEvent)
					if commitErr := storeOffsets(consumer, msg.TopicPartition); commitErr != nil {
						ConsumerLogger.Error().Err(commitErr).Msg("offset commit error")
					}

				}
			}
		}
	}(routerCtx, consumer, instanceID, wg)
}

func StartConsumers(routerCtx context.Context, applicationContext ApplicationContext, consumerConfig *kafka.ConfigMap, topicEntity string, topics []string, instances int, handlerFunc HandlerFunc, wg *sync.WaitGroup) []*kafka.Consumer {
	consumers := make([]*kafka.Consumer, 0, instances)
	for i := 0; i < instances; i++ {
		consumer := createConsumer(consumerConfig, topics)
		consumers = append(consumers, consumer)
		groupID, _ := consumerConfig.Get("group.id", "")
		instanceID := fmt.Sprintf("%s-%s-%d", topicEntity, groupID, i)
		wg.Add(1)
		startConsumer(routerCtx, applicationContext, handlerFunc, consumer, topicEntity, instanceID, wg)
	}
	return consumers
}
