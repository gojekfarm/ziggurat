package rabbitmq

import "github.com/gojekfarm/ziggurat"

type Opts func(r *autoRetry)

func WithHosts(hosts []string) Opts {
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

type publishOpts struct {
	retryCount      int
	queueKey        string
	delayExpiration string
}

type PublishOpts func(po *publishOpts)

func WithRetryCount(count int) PublishOpts {
	return func(p *publishOpts) {
		p.retryCount = count
	}
}

func WithQueue(queue string) PublishOpts {
	return func(po *publishOpts) {
		po.queueKey = queue
	}
}

func WithDelayExpiration(expiration string) PublishOpts {
	return func(po *publishOpts) {
		po.delayExpiration = expiration
	}
}
