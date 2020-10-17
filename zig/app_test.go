package zig

import (
	"github.com/julienschmidt/httprouter"
	"testing"
)

var app = &App{}

type mockHTTP struct {
	isStartInvoked bool
}
type mockStatsD struct {
	isStartInvoked bool
}

type mockRouter struct {
	isStartInvoked bool
}

type mockRabbitMQ struct {
	isStartInvoked bool
}

func (m *mockRabbitMQ) Start(app *App) (chan int, error) {
	m.isStartInvoked = true
	stopChan := make(chan int)
	go func() {
		close(stopChan)
	}()
	return stopChan, nil
}

func (m *mockRabbitMQ) Retry(app *App, payload MessageEvent) error {
	return nil
}

func (m *mockRabbitMQ) Stop() error {
	return nil
}

func (m *mockRabbitMQ) Replay(app *App, topicEntity string, count int) error {
	return nil
}

func (m *mockStatsD) Start(app *App) error {
	m.isStartInvoked = true
	return nil
}

func (m *mockStatsD) Stop() error {
	return nil
}

func (m *mockStatsD) IncCounter(metricName string, value int, arguments map[string]string) error {
	return nil
}

func (m *mockRouter) Start(app *App) (chan int, error) {
	m.isStartInvoked = true
	closeChan := make(chan int)
	go func() {
		close(closeChan)
	}()
	return closeChan, nil
}

func (m *mockRouter) HandlerFunc(topicEntityName string, handlerFn HandlerFunc, mw Middleware) {

}

func (m *mockRouter) GetTopicEntities() []*topicEntity {
	return []*topicEntity{}
}

func (m *mockRouter) GetHandlerFunctionMap() map[string]*topicEntity {
	return map[string]*topicEntity{}
}

func (mh *mockHTTP) Start(app *App) {
	mh.isStartInvoked = true
}

func (mh *mockHTTP) attachRoute(func(r *httprouter.Router)) {

}

func (mh *mockHTTP) Stop() error {
	return nil
}

func TestApp_Start(t *testing.T) {
	http, router, statsd, rabbitmq := &mockHTTP{}, &mockRouter{}, &mockStatsD{}, &mockRabbitMQ{}
	startCallbackCalled := false
	app.Router = router
	app.HttpServer = http
	app.MetricPublisher = statsd
	app.Retrier = rabbitmq
	app.cancelFun = func() {}
	startCallback := func(app *App) {
		startCallbackCalled = true
	}

	app.start(startCallback)

	if !(http.isStartInvoked && router.isStartInvoked && statsd.isStartInvoked && rabbitmq.isStartInvoked && startCallbackCalled) {
		t.Errorf("expected %v got %v", true, false)
	}
}
