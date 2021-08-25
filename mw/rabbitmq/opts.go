package rabbitmq

import "github.com/gojekfarm/ziggurat"

type Opts func(r *autoRetry)

func WithHosts(hosts ...string) Opts {
	return func(r *autoRetry) {
		r.hosts = hosts
	}
}

func WithUsername(username string) Opts {
	return func(r *autoRetry) {
		r.username = username
	}
}

func WithPassword(password string) Opts {
	return func(r *autoRetry) {
		r.password = password
	}
}

func WithLogger(l ziggurat.StructuredLogger) Opts {
	return func(r *autoRetry) {
		r.logger = &amqpExtraLogger{
			l: l,
		}
		r.ogLogger = l
	}
}
