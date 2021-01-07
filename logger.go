package ziggurat

type StructuredLogger interface {
	Info(message string, kvs ...map[string]interface{})
	Debug(message string, kvs ...map[string]interface{})
	Warn(message string, kvs ...map[string]interface{})
	Error(message string, err error, kvs ...map[string]interface{})
	Fatal(message string, err error, kvs ...map[string]interface{})
}
