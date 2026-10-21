package rabbitmq

import (
	"reflect"
	"testing"

	"github.com/streadway/amqp"
)

func TestAmqpHeadersToMap(t *testing.T) {
	t.Run("nil for no headers", func(t *testing.T) {
		if got := amqpHeadersToMap(nil); got != nil {
			t.Errorf("expected nil for no headers, got %v", got)
		}
		if got := amqpHeadersToMap(amqp.Table{}); got != nil {
			t.Errorf("expected nil for empty headers, got %v", got)
		}
	})

	t.Run("converts string and non-string values", func(t *testing.T) {
		headers := amqp.Table{
			"traceparent":  "00-abc-def-01",
			"retry-origin": "ziggurat-go",
			"retry-count":  3,
		}
		got := amqpHeadersToMap(headers)
		want := map[string]string{
			"traceparent":  "00-abc-def-01",
			"retry-origin": "ziggurat-go",
			"retry-count":  "3",
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("expected %v, got %v", want, got)
		}
	})
}
