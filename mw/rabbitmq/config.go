package rabbitmq

type QueueConfig struct {
	QueueName             string
	DelayExpirationInMS   string
	RetryCount            int
	WorkerCount           int
	ConsumerPrefetchCount int
}
