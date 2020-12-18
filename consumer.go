package ziggurat

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"sync"
	"time"
)

const defaultPollTimeout = 100 * time.Millisecond
const brokerRetryTimeout = 2 * time.Second

var startConsumer = func(z *Ziggurat, h MessageHandler, consumer *kafka.Consumer, route string, instanceID string, wg *sync.WaitGroup) {
	go func(instanceID string) {
		defer wg.Done()
		doneCh := z.Context().Done()
		worker := NewWorker(10)
		sendCh, _ := worker.run(z, func(message *kafka.Message) {
			processor(message, route, consumer, h, z)
		})
		for {
			select {
			case <-doneCh:
				close(sendCh)
				return
			default:
				msg, err := readMessage(consumer, defaultPollTimeout)
				if err != nil && err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				} else if err != nil && err.(kafka.Error).Code() == kafka.ErrAllBrokersDown {
					LogError(err, "retrying broker...", nil)
					time.Sleep(brokerRetryTimeout)
					continue
				}
				if msg != nil {
					sendCh <- msg
				}
			}
		}
	}(instanceID)
}

var StartConsumers = func(z *Ziggurat, consumerConfig *kafka.ConfigMap, route string, topics []string, instances int, h MessageHandler, wg *sync.WaitGroup) []*kafka.Consumer {
	consumers := make([]*kafka.Consumer, 0, instances)
	for i := 0; i < instances; i++ {
		consumer := createConsumer(consumerConfig, topics)
		consumers = append(consumers, consumer)
		groupID, _ := consumerConfig.Get("group.id", "")
		instanceID := fmt.Sprintf("%s_%s_%d", route, groupID, i)
		wg.Add(1)
		startConsumer(z, h, consumer, route, instanceID, wg)
	}
	return consumers
}
