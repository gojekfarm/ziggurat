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
