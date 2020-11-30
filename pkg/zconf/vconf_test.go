package zconf

import (
	"errors"
	"github.com/gojekfarm/ziggurat-go/pkg/zb"
	"os"
	"reflect"
	"testing"
)

const testConfPath = "../../config/config.test.yaml"

func TestViperConfig_Parse(t *testing.T) {
	vc := NewViperConfig()
	expectedConfig := zb.Config{
		StreamRouter: map[string]zb.StreamRouterConfig{
			"plain-text-log": {
				InstanceCount:    2,
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				GroupID:          "plain_text_consumer",
			},
		},
		LogLevel:    "debug",
		ServiceName: "test-app",
		Retry: zb.RetryConfig{
			Enabled: true,
			Count:   5,
		},
		HTTPServer: zb.HTTPServerConfig{
			Port: "8080",
		},
	}
	vc.Parse(zb.CommandLineOptions{ConfigFilePath: testConfPath})
	actualConfig := vc.Config()
	if !reflect.DeepEqual(expectedConfig, *actualConfig) {
		t.Errorf("expected config %+v, actual cfgReader %+v", expectedConfig, actualConfig)
	}

}

func TestViperConfig_EnvOverride(t *testing.T) {
	overriddenValue := "localhost:9094"
	vc := NewViperConfig()
	if err := os.Setenv("ZIGGURAT_STREAM_ROUTER_PLAIN_TEXT_LOG_BOOTSTRAP_SERVERS", overriddenValue); err != nil {
		t.Error(err)
	}
	vc.Parse(zb.CommandLineOptions{ConfigFilePath: testConfPath})
	config := vc.Config()
	actualValue := config.StreamRouter["plain-text-log"].BootstrapServers
	if !(actualValue == overriddenValue) {
		t.Errorf("expected value of bootstrap servers to be %s but got %s", overriddenValue, actualValue)
	}
}

func TestViperConfig_Validate(t *testing.T) {
	vc := NewViperConfig()
	vc.ParsedConfig = &zb.Config{
		StreamRouter: nil,
		LogLevel:     "",
		ServiceName:  "foo",
		Retry:        zb.RetryConfig{},
		HTTPServer:   zb.HTTPServerConfig{},
	}

	validationError := errors.New("service cannot be foo")

	rules := map[string]func(c *zb.Config) error{
		"serviceNameValidation": func(c *zb.Config) error {
			if c.ServiceName == "foo" {
				return validationError
			}
			return nil
		},
	}

	err := vc.Validate(rules)
	if err == nil {
		t.Errorf("expected error to be %v, got %v", validationError, err)
	}
}

func TestViperConfig_GetByKey(t *testing.T) {
	vc := NewViperConfig()
	vc.Parse(zb.CommandLineOptions{ConfigFilePath: testConfPath})
	expectedStatsDConf := map[string]interface{}{"host": "localhost:8125"}
	statsdCfg := vc.GetByKey("statsd").(map[string]interface{})

	if !reflect.DeepEqual(expectedStatsDConf, statsdCfg) {
		t.Errorf("expected %v got %v", expectedStatsDConf, statsdCfg)
	}
}
