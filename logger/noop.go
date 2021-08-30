package logger

type noopLogger struct{}

func (n noopLogger) Info(message string, kvs ...map[string]interface{}) {

}

func (n noopLogger) Debug(message string, kvs ...map[string]interface{}) {
}

func (n noopLogger) Warn(message string, kvs ...map[string]interface{}) {
}

func (n noopLogger) Error(message string, err error, kvs ...map[string]interface{}) {
}

func (n noopLogger) Fatal(message string, err error, kvs ...map[string]interface{}) {
}

var NoopLogger = noopLogger{}
