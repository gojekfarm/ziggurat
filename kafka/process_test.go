package kafka

import (
	"context"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/logger"
)

func Test_processMessage(t *testing.T) {
	dl := logger.NewDiscardLogger()
	var oldStoreOffsets = storeOffsets
	defer func() {
		storeOffsets = oldStoreOffsets
	}()
	storeOffsets = func(consumer *kafka.Consumer, partition kafka.TopicPartition) error {
		return nil
	}
	topic := "bar"
	cfmMap := kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "foo",
	}
	part := 1
	wantPath := "localhost:9092/foo/bar/1"

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
	processMessage(context.Background(), &msg, &kafka.Consumer{}, h, dl, cfmMap, "")

}
