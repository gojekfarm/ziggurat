package rabbitmq

import "github.com/gojekfarm/ziggurat/v2"

const KeyRetryCount = "rabbitmqAutoRetryCount"

func RetryCountFor(e *ziggurat.Event) int {
	if e.Metadata == nil {
		return 0
	}

	c, _ := e.Metadata[KeyRetryCount]

	switch count := c.(type) {
	case int:
		return count
		//interface{} which is a number is unmarshalled as a float64
	case float64:
		return int(count)
	default:
		return 0
	}
}
