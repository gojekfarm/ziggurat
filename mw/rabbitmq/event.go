package rabbitmq

import "github.com/gojekfarm/ziggurat"

const RetryCountKey = "retryCount"

func getRetryCount(e *ziggurat.Event) int {
	c, ok := e.Metadata[RetryCountKey]
	if !ok {
		return 0
	}
	count, ok := c.(int)
	if !ok {
		return 0
	}
	return count
}
