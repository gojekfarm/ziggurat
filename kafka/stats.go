package kafka

type Stats interface {
	Inc(name string, value int, options ...interface{})
	Gauge(name string, value float64, options ...interface{})
}
