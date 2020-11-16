package basic

type CommandLineOptions struct {
	ConfigFilePath string
}

type StreamRouterConfig struct {
	InstanceCount    int    `mapstructure:"instance-count"`
	BootstrapServers string `mapstructure:"bootstrap-servers"`
	OriginTopics     string `mapstructure:"origin-topics"`
	GroupID          string `mapstructure:"group-id"`
}

type RetryConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Count   int  `mapstructure:"count"`
}

type HTTPServerConfig struct {
	Port string `mapstructure:"port"`
}

type Config struct {
	StreamRouter map[string]StreamRouterConfig `mapstructure:"stream-router"`
	LogLevel     string                        `mapstructure:"log-level"`
	ServiceName  string                        `mapstructure:"service-name"`
	Retry        RetryConfig                   `mapstructure:"retry"`
	HTTPServer   HTTPServerConfig              `mapstructure:"http-server"`
}
