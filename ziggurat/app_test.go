package ziggurat

import (
	"github.com/julienschmidt/httprouter"
	"testing"
)

type mockHTTP struct{}
type mockStatsD struct{}
type mockRouter struct{}

func (m *mockRouter) GetTopicEntityNames() []string {
	return []string{}
}

type mockRabbitMQ struct{}

var app *Ziggurat
var mhttp, mrouter, mstatsd, mrabbitmq = &mockHTTP{}, &mockRouter{}, &mockStatsD{}, &mockRabbitMQ{}
var startCount = 0
var stopCount = 0
var expectedStopCount = 3
var expectedStartCount = 4

func (m *mockRabbitMQ) Start(app *Ziggurat) (chan int, error) {
	startCount++
	stopChan := make(chan int)
	go func() {
		close(stopChan)
	}()
	return stopChan, nil
}

func (m *mockRabbitMQ) Retry(app *Ziggurat, payload MessageEvent) error {
	return nil
}

func (m *mockRabbitMQ) Stop() error {
	stopCount++
	return nil
}

func (m *mockRabbitMQ) Replay(app *Ziggurat, topicEntity string, count int) error {
	return nil
}

func (m *mockStatsD) Start(app *Ziggurat) error {
	startCount++
	return nil
}

func (m *mockStatsD) Stop() error {
	stopCount++
	return nil
}

func (m *mockStatsD) Gauge(metricName string, value int64, arguments map[string]string) error {
	return nil
}

func (m *mockStatsD) IncCounter(metricName string, value int64, arguments map[string]string) error {
	return nil
}

func (m *mockRouter) Start(app *Ziggurat) (chan int, error) {
	startCount++
	closeChan := make(chan int)
	go func() {
		close(closeChan)
	}()
	return closeChan, nil
}

func (m *mockRouter) HandlerFunc(topicEntityName string, handlerFn HandlerFunc, mw ...MiddlewareFunc) {

}

func (m *mockRouter) GetTopicEntities() []*topicEntity {
	return []*topicEntity{}
}

func (m *mockRouter) GetHandlerFunctionMap() map[string]*topicEntity {
	return map[string]*topicEntity{}
}

func (mh *mockHTTP) Start(app *Ziggurat) {
	startCount++
}

func (mh *mockHTTP) DefineRoutes(func(r *httprouter.Router)) {

}

func (mh *mockHTTP) Stop() error {
	stopCount++
	return nil
}

func (mh *mockHTTP) ConfigureHTTPRoutes(a *Ziggurat, configFunc func(a *Ziggurat, r *httprouter.Router)) {
	panic("implement me")
}

func setup() {
	app = &Ziggurat{}
	app.router = mrouter
	app.httpServer = mhttp
	app.metricPublisher = mstatsd
	app.messageRetry = mrabbitmq
	app.cancelFun = func() {}
	app.config = &Config{
		StreamRouter: nil,
		LogLevel:     "",
		ServiceName:  "",
		Retry:        RetryConfig{Enabled: true},
		HTTPServer:   HTTPServerConfig{},
	}
}

func teardown() {
	app = &Ziggurat{}
	startCount = 0
	stopCount = 0
}

func TestApp_Start(t *testing.T) {
	setup()
	defer teardown()
	startCallbackCalled := false
	startCallback := func(app *Ziggurat) {
		startCallbackCalled = true
	}

	app.start(startCallback, nil)

	if startCount < expectedStartCount {
		t.Errorf("expected start count to be %v but got %v", expectedStartCount, startCount)
	}

	if !startCallbackCalled {
		t.Errorf("expected startCallbackCalled to be %v, but got %v", true, startCallbackCalled)
	}

}

func TestApp_Stop(t *testing.T) {
	setup()
	defer teardown()
	stopCallbackCalled := false

	app.stop(func() {
		stopCallbackCalled = true
	})
	if stopCount < expectedStopCount {
		t.Errorf("expected stop count to be %v, but got %v", expectedStopCount, stopCount)
	}
	if !stopCallbackCalled {
		t.Errorf("expected stopCallbackCalled to be %v, but got %v", true, stopCallbackCalled)
	}
}