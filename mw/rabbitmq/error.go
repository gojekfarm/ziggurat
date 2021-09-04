package rabbitmq

type retryable interface {
	Retry() bool
}

type publishFailed interface {
	PublishFailed() bool
}

func isRetryAble(e error) bool {
	re, ok := e.(retryable)
	return ok && re.Retry()
}

func isPublishFailed(e error) bool {
	pf, ok := e.(publishFailed)
	return ok && pf.PublishFailed()
}

type errRetry struct {
	e error
}

func (err errRetry) Error() string {
	return err.e.Error()
}

func NewErrRetry(err error) errRetry {
	return errRetry{e: err}
}

func (err errRetry) PublishFailed() bool {
	if err.e == errRabbitMQPublish {
		return true
	}
	return false
}

func (err errRetry) Retry() bool {
	return true
}
