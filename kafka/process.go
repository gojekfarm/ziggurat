package kafka

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
)

const (
	HeaderTopic     = "x-kafka-topic"
	HeaderPartition = "x-kafka-partition"
	EventType       = "kafka"
)

func constructPath(bs string, cg string, topic string, part int32) string {
	return fmt.Sprintf("%s/%s/%s/%d", bs, cg, topic, part)
}

func getStrValFromCfgMap(cfgMap kafka.ConfigMap, prop string) string {
	//we ignore the error as the createConsumer panics if
	// bootstrap.servers is not present
	v, _ := cfgMap.Get(prop, "")
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// processMessage executes the handler for every message that is received
// all metadata is serialized to strings and set in headers
func processMessage(ctx context.Context,
	msg *kafka.Message,
	c *kafka.Consumer,
	h ziggurat.Handler,
	l ziggurat.StructuredLogger,
	cfgMap kafka.ConfigMap,
	route string) {

	//we ignore the errors as createConsumer panics if
	// bootstrap.servers is not present
	bs := getStrValFromCfgMap(cfgMap, "bootstrap.servers")
	cg := getStrValFromCfgMap(cfgMap, "group.id")

	event := ziggurat.Event{
		Headers: map[string]string{
			HeaderPartition: strconv.Itoa(int(msg.TopicPartition.Partition)),
			HeaderTopic:     *msg.TopicPartition.Topic,
		},
		Value:             msg.Value,
		Key:               msg.Key,
		Path:              route,
		Metadata:          map[string]interface{}{},
		RoutingPath:       constructPath(bs, cg, *msg.TopicPartition.Topic, msg.TopicPartition.Partition),
		ProducerTimestamp: msg.Timestamp,
		ReceivedTimestamp: time.Now(),
		EventType:         EventType,
	}
	l.Info(fmt.Sprintf("%s processing message", c.String()))
	l.Error("kafka processing error", h.Handle(ctx, &event))
	err := storeOffsets(c, msg.TopicPartition)
	l.Error("error storing offsets: %v", err)
}
