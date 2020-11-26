package server

import (
	"context"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"net"
	"testing"
)

const serverAddr = "localhost:8080"

type testConfigStore struct{}
type testMockApp struct{}

func (t testMockApp) Context() context.Context {
	panic("implement me")
}

func (t testMockApp) Routes() []string {
	panic("implement me")
}

func (t testMockApp) MessageRetry() z.MessageRetry {
	panic("implement me")
}

func (t testMockApp) Handler() z.MessageHandler {
	panic("implement me")
}

func (t testMockApp) MetricPublisher() z.MetricPublisher {
	panic("implement me")
}

func (t testMockApp) HTTPServer() z.HttpServer {
	panic("implement me")
}

func (t testMockApp) Config() *basic.Config {
	panic("implement me")
}

func (t testMockApp) ConfigStore() z.ConfigStore {
	panic("implement me")
}

func (t testConfigStore) Config() *basic.Config {
	return &basic.Config{
		HTTPServer: basic.HTTPServerConfig{
			Port: "8080",
		},
	}
}

func (t testConfigStore) Parse(options basic.CommandLineOptions) {
	panic("implement me")
}

func (t testConfigStore) GetByKey(key string) interface{} {
	panic("implement me")
}

func (t testConfigStore) Validate(rules map[string]func(c *basic.Config) error) error {
	panic("implement me")
}

func (t testConfigStore) UnmarshalByKey(key string, model interface{}) error {
	panic("implement me")
}

func TestDefaultHttpServer_Start(t *testing.T) {
	ds := NewDefaultHTTPServer(&testConfigStore{})
	ds.Start(testMockApp{})
	_, err := net.Dial("tcp", serverAddr)
	if err != nil {
		t.Errorf("error connection to server %s", err)
	}
	ds.Stop(testMockApp{})
}
