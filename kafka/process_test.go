package kafka

import (
	"context"
	"fmt"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/logger"
)

func Test_processMessage(t *testing.T) {

	dl := logger.NOOP
	var oldStoreOffsets = storeOffsets
	defer func() {
		storeOffsets = oldStoreOffsets
	}()
	storeOffsets = func(consumer *kafka.Consumer, partition kafka.TopicPartition) error {
		return nil
	}
	topic := "bar"
	route := "baz"

	part := 1
	wantPath := fmt.Sprintf("%s/%s/%d", route, topic, part)

	msg := kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: int32(part),
		},
	}
	h := ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) error {
		if event.RoutingPath != wantPath {
			t.Errorf("expecting path to be %s but got %s", wantPath, event.RoutingPath)
		}
		return nil
	})
	processMessage(context.Background(), &msg, h, dl, route)

}
