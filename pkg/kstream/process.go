package kstream

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat/pkg/z"
	"github.com/gojekfarm/ziggurat/pkg/zb"
	"github.com/gojekfarm/ziggurat/pkg/zlog"
)

func processor(msg *kafka.Message, route string, c *kafka.Consumer, h z.MessageHandler, a z.App) {
	event := zb.NewMessageEvent(msg.Key, msg.Value, *msg.TopicPartition.Topic, route, msg.TimestampType.String(), msg.Timestamp)
	h.HandleMessage(event, a)
	zlog.LogError(storeOffsets(c, msg.TopicPartition), "consumer error", nil)
}
