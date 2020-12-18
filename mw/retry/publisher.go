package retry

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/gojekfarm/ziggurat"
)

func constructQueueName(routeName StreamRouteName, queueType string) string {
	return fmt.Sprintf("%s_%s_%s_queue", "ziggurat", routeName, queueType)
}

func constructExchangeName(route StreamRouteName, queueType string) string {
	return fmt.Sprintf("%s_%s_%s_exchange", "ziggurat", route, queueType)
}

func getRetryCount(m *ziggurat.MessageEvent) int {
	if value := m.GetMessageAttribute("retryCount"); value == nil {
		return 0
	}
	return m.GetMessageAttribute("retryCount").(int)
}

func setRetryCount(m *ziggurat.MessageEvent) {
	value := m.GetMessageAttribute("retryCount")

	if value == nil {
		m.SetMessageAttribute("retryCount", 1)
		return
	}
	m.SetMessageAttribute("retryCount", value.(int)+1)
}

func encodeMessage(message ziggurat.MessageEvent) (*bytes.Buffer, error) {
	buff := bytes.NewBuffer([]byte{})
	encoder := gob.NewEncoder(buff)

	if err := encoder.Encode(message); err != nil {
		return nil, err
	}
	return buff, nil
}
