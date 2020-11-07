package zig

import (
	"os"
	"reflect"
	"testing"
)

func TestConfig_parseConfig(t *testing.T) {
	parseConfig(CommandLineOptions{ConfigFilePath: "../config/config.test.yaml"})
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
	actualConfig := getConfig()
	if !reflect.DeepEqual(expectedConfig, actualConfig) {
		t.Errorf("expected confg %+v, actual config %+v", expectedConfig, actualConfig)
	}
}

func TestEnvOverride(t *testing.T) {
	overriddenValue := "localhost:9094"
	if err := os.Setenv("ZIGGURAT_STREAM_ROUTER_PLAIN_TEXT_LOG_BOOTSTRAP_SERVERS", overriddenValue); err != nil {
		t.Error(err)
	}
	parseConfig(CommandLineOptions{ConfigFilePath: "../config/config.test.yaml"})
	config := getConfig()
	actualValue := config.StreamRouter["plain-text-log"].BootstrapServers
	if !(actualValue == overriddenValue) {
		t.Errorf("expected value of bootstrap servers to be %s but got %s", overriddenValue, actualValue)
	}

}
