package kstream

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/zlog"
	"github.com/gojekfarm/ziggurat/ztype"
)

func processor(msg *kafka.Message, route string, c *kafka.Consumer, h ztype.MessageHandler, a ztype.App) {
	event := zbase.NewMessageEvent(msg.Key, msg.Value, *msg.TopicPartition.Topic, route, msg.TimestampType.String(), msg.Timestamp)
	h.HandleMessage(event, a)
	zlog.LogError(storeOffsets(c, msg.TopicPartition), "consumer error", nil)
}
