package kstream

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zbasic"
	"github.com/gojekfarm/ziggurat-go/pkg/zlogger"
)

func processor(msg *kafka.Message, route string, c *kafka.Consumer, h z.MessageHandler, a z.App) {
	event := zbasic.NewMessageEvent(msg.Key, msg.Value, *msg.TopicPartition.Topic, route, msg.TimestampType.String(), msg.Timestamp)
	h.HandleMessage(event, a)
	zlogger.LogError(storeOffsets(c, msg.TopicPartition), "consumer error", nil)
}
