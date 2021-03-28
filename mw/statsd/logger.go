package statsd

type noopLogger func()

func (n noopLogger) Info(message string, kvs ...map[string]interface{}) {
	n()
}

func (n noopLogger) Debug(message string, kvs ...map[string]interface{}) {
	n()
}

func (n noopLogger) Warn(message string, kvs ...map[string]interface{}) {
	n()
}

func (n noopLogger) Error(message string, err error, kvs ...map[string]interface{}) {
	n()
}

func (n noopLogger) Fatal(message string, err error, kvs ...map[string]interface{}) {
	n()
}
