package kafka

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/gojekfarm/ziggurat/v2"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	EventType = "kafka"
)

func constructPath(rg string, topic string, part int32) string {
	return fmt.Sprintf("%s/%s/%d", rg, topic, part)
}

// processMessage builds a ziggurat.Event from a raw kafka.Message and invokes the
// handler. A panic inside the handler is recovered here and logged rather than being
// allowed to propagate: an unrecovered panic in a worker goroutine would otherwise
// terminate the entire process, and would also skip the wg.Done() call the caller
// relies on to know this worker has finished, permanently deadlocking
// ConsumerGroup.Consume's wg.Wait() for every other worker in the group.
func processMessage(ctx context.Context, msg *kafka.Message, h ziggurat.Handler, route string, logger ziggurat.StructuredLogger) {
	//copy kvs into new slices
	key := make([]byte, len(msg.Key))
	value := make([]byte, len(msg.Value))

	copy(key, msg.Key)
	copy(value, msg.Value)

	event := ziggurat.Event{
		Value: value,
		Key:   key,
		Metadata: map[string]interface{}{
			"kafka-topic":     *msg.TopicPartition.Topic,
			"kafka-partition": int(msg.TopicPartition.Partition),
		},
		RoutingPath:       constructPath(route, *msg.TopicPartition.Topic, msg.TopicPartition.Partition),
		ProducerTimestamp: msg.Timestamp,
		ReceivedTimestamp: time.Now(),
		EventType:         EventType,
	}

	defer func() {
		if r := recover(); r != nil {
			if logger != nil {
				logger.Error(
					"recovered from panic in handler",
					fmt.Errorf("panic: %v", r),
					map[string]interface{}{
						"routing_path": event.RoutingPath,
						"kafka-topic":  event.Metadata["kafka-topic"],
						"stack":        string(debug.Stack()),
					},
				)
			}
		}
	}()

	h.Handle(ctx, &event)
}
