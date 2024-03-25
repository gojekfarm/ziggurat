package rabbitmq

import (
	"github.com/gojekfarm/ziggurat/v2"
	"testing"
)

func Test_getRetryCount(t *testing.T) {
	type test struct {
		name     string
		input    *ziggurat.Event
		expected int
	}

	tests := []test{
		{
			name:     "ARetry count should be 0 when metadata is nil",
			input:    &ziggurat.Event{Metadata: nil},
			expected: 0,
		},
		{
			name:     "ARetry count when metadata does not contain retryCount key",
			input:    &ziggurat.Event{Metadata: map[string]interface{}{}},
			expected: 0,
		},
		{
			name:     "ARetry count when key is present",
			input:    &ziggurat.Event{Metadata: map[string]interface{}{KeyRetryCount: 5}},
			expected: 5,
		}}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			count := RetryCountFor(c.input)
			if count != c.expected {
				t.Errorf("expected count to be %d got %d", c.expected, count)
			}
		})
	}
}
