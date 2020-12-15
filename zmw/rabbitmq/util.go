package rabbitmq

import (
	"fmt"
	"github.com/gojekfarm/ziggurat/zbase"
)

func constructQueueName(prefix string, topicEntity string, queueType string) string {
	return fmt.Sprintf("%s_%s_%s_queue", topicEntity, prefix, queueType)
}

func constructExchangeName(serviceName string, topicEntity string, exchangeType string) string {
	return fmt.Sprintf("%s_%s_%s_exchange", topicEntity, serviceName, exchangeType)
}

func getRetryCount(m *zbase.MessageEvent) int {
	if value := m.GetMessageAttribute("retryCount"); value == nil {
		return 0
	}
	return m.GetMessageAttribute("retryCount").(int)
}

func setRetryCount(m *zbase.MessageEvent) {
	value := m.GetMessageAttribute("retryCount")

	if value == nil {
		m.SetMessageAttribute("retryCount", 1)
		return
	}
	m.SetMessageAttribute("retryCount", value.(int)+1)
}
