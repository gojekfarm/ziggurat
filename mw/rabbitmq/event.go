package rabbitmq

import (
	"github.com/gojekfarm/ziggurat"
)

const KeyRetryCount = "retryCount"

func getRetryCount(e *ziggurat.Event) int {
	if e.Metadata == nil {
		return 0
	}

	c, ok := e.Metadata[KeyRetryCount]
	if !ok {
		return 0
	}

	count, ok := c.(float64)
	if !ok {
		panic("couldn't assert count to a number")
	}

	return int(count)
}
