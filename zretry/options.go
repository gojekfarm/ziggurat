package zretry

import (
	"strings"
)

type RabbitMQOptions func(r *RabbitMQConfig)

func WithHosts(urls ...string) RabbitMQOptions {
	return func(r *RabbitMQConfig) {
		r.hosts = strings.Join(urls, ",")
	}
}

func WithQueuePrefix(prefix string) RabbitMQOptions {
	return func(r *RabbitMQConfig) {
		r.queuePrefix = prefix
	}
}

func WithDelayQueueExpiration(expirationInMS string) RabbitMQOptions {
	return func(r *RabbitMQConfig) {
		r.delayQueueExpiration = expirationInMS
	}
}

func WithDialTimeoutInSeconds(timeout int) RabbitMQOptions {
	return func(r *RabbitMQConfig) {
		r.dialTimeoutInS = timeout
	}
}

func WithRetryCount(count int) RabbitMQOptions {
	return func(r *RabbitMQConfig) {
		r.count = count
	}
}
