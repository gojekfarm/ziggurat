package rabbitmq

import (
	"github.com/gojekfarm/ziggurat"
)

const KeyRetryCount = "retryCount"

func getRetryCount(e *ziggurat.Event) int {
	if e.Metadata == nil {
		return 0
	}

	c, _ := e.Metadata[KeyRetryCount]

	switch count := c.(type) {
	case int:
		return count
	case float64:
		return int(count)
	default:
		return 0
	}
}
