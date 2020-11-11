package zig

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestViperConfig_Parse(t *testing.T) {
	vc := NewViperConfig()
	expectedConfig := Config{
		StreamRouter: map[string]StreamRouterConfig{
			"plain-text-log": {
				InstanceCount:    2,
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				GroupID:          "plain_text_consumer",
			},
		},
		LogLevel:    "debug",
		ServiceName: "test-app",
		Retry: RetryConfig{
			Enabled: true,
			Count:   5,
		},
		HTTPServer: HTTPServerConfig{
			Port: "8080",
		},
	}
	vc.Parse(CommandLineOptions{ConfigFilePath: "../config/config.test.yaml"})
	actualConfig := vc.Config()
	if !reflect.DeepEqual(expectedConfig, *actualConfig) {
		t.Errorf("expected config %+v, actual appconf %+v", expectedConfig, actualConfig)
	}

}

func TestViperConfig_EnvOverride(t *testing.T) {
	overriddenValue := "localhost:9094"
	vc := NewViperConfig()
	if err := os.Setenv("ZIGGURAT_STREAM_ROUTER_PLAIN_TEXT_LOG_BOOTSTRAP_SERVERS", overriddenValue); err != nil {
		t.Error(err)
	}
	vc.Parse(CommandLineOptions{ConfigFilePath: "../config/config.test.yaml"})
	config := vc.Config()
	actualValue := config.StreamRouter["plain-text-log"].BootstrapServers
	if !(actualValue == overriddenValue) {
		t.Errorf("expected value of bootstrap servers to be %s but got %s", overriddenValue, actualValue)
	}
}

func TestViperConfig_Validate(t *testing.T) {
	vc := NewViperConfig()
	vc.parsedConfig = &Config{
		StreamRouter: nil,
		LogLevel:     "",
		ServiceName:  "foo",
		Retry:        RetryConfig{},
		HTTPServer:   HTTPServerConfig{},
	}

	validationError := errors.New("service cannot be foo")

	rules := map[string]func(c *Config) error{
		"serviceNameValidation": func(c *Config) error {
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
	vc.Parse(CommandLineOptions{ConfigFilePath: "../config/config.test.yaml"})
	expectedStatsDConf := map[string]interface{}{"host": "localhost:8125"}
	statsdCfg := vc.GetByKey("statsd").(map[string]interface{})

	if !reflect.DeepEqual(expectedStatsDConf, statsdCfg) {
		t.Errorf("expected %v got %v", expectedStatsDConf, statsdCfg)
	}

}
