package rabbitmq

import (
	"github.com/gojekfarm/ziggurat/v2"
	"time"
)

type Opts func(r *ARetry)

func WithHosts(hosts ...string) Opts {
	return func(r *ARetry) {
		r.hosts = hosts
	}
}

func WithUsername(username string) Opts {
	return func(r *ARetry) {
		r.username = username
	}
}

func WithPassword(password string) Opts {
	return func(r *ARetry) {
		r.password = password
	}
}

func WithLogger(l ziggurat.StructuredLogger) Opts {
	return func(r *ARetry) {
		r.logger = &amqpExtraLogger{
			l: l,
		}
		r.ogLogger = l
	}
}

func WithConnectionTimeout(t time.Duration) Opts {
	return func(r *ARetry) {
		r.connTimeout = t
	}
}

// defaultPublisherConfirmsBuffer is used when publisher confirms are enabled via
// WithPublisherConfirms(0) - the underlying amqpextra publisher requires a buffer of
// at least 1, and 0 is also the zero value a caller would get by forgetting to set it,
// so treat it as "use a sensible default" rather than a configuration error.
const defaultPublisherConfirmsBuffer = 100

// WithPublisherConfirms turns on RabbitMQ publisher confirms (see
// https://www.rabbitmq.com/docs/confirms#publisher-confirms) for every publisher this
// ARetry creates - the ones used by Publish, Retry (and the DLQ delay/instant republish
// it triggers), and DSReplayHandler's replay publisher.
//
// With confirms on, the broker acknowledges (or negatively acknowledges) each published
// message once it's been handled according to the queue's durability guarantees; Publish
// and Retry's returned error will then reflect a broker-level nack, not just a local
// write-to-socket success. Without this option (the default), a nil error only means the
// message was written to the TCP connection - it says nothing about whether the broker
// actually accepted and routed it.
//
// This does add latency to every publish call, since it now waits for the broker's
// acknowledgement instead of returning as soon as the message is handed to the channel.
//
// buffer controls how many confirms can be in-flight at once (how many messages may be
// awaiting acknowledgement concurrently across all publishers in the pool) - once that
// limit is reached, further publishes block until an earlier one is confirmed. Pass 0 to
// use a default of 100, which is generous for typical retry/DLQ volumes.
func WithPublisherConfirms(buffer uint) Opts {
	if buffer == 0 {
		buffer = defaultPublisherConfirmsBuffer
	}
	return func(r *ARetry) {
		r.publisherConfirms = true
		r.publisherConfirmsBuffer = buffer
	}
}
