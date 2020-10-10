package ziggurat

type MetricPublisher interface {
	Start(app *App) error
	Stop() error
	IncCounter(metricName string, value int, arguments map[string]string) error
}
