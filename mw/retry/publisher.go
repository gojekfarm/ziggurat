package retry

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/gojekfarm/ziggurat"
	"time"
)

func constructQueueName(routeName StreamRouteName, queueType string) string {
	return fmt.Sprintf("%s_%s_%s_queue", "ziggurat", routeName, queueType)
}

func constructExchangeName(route StreamRouteName, queueType string) string {
	return fmt.Sprintf("%s_%s_%s_exchange", "ziggurat", route, queueType)
}

func getRetryCount(m *ziggurat.Message) int {
	if value := m.Attribute("retryCount"); value == nil {
		return 0
	}
	return m.Attribute("retryCount").(int)
}

func setRetryCount(m *ziggurat.Message) {
	value := m.Attribute("retryCount")

	if value == nil {
		m.SetAttribute("retryCount", 1)
		return
	}
	m.SetAttribute("retryCount", value.(int)+1)
}

func encodeMessage(message *ziggurat.Message) (*bytes.Buffer, error) {
	buff := bytes.NewBuffer([]byte{})
	gob.Register(time.Time{})
	encoder := gob.NewEncoder(buff)

	if err := encoder.Encode(message); err != nil {
		return nil, err
	}
	return buff, nil
}
