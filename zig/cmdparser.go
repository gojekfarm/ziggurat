package zig

import "flag"

var (
	config string
)

func init() {
	flag.StringVar(&config, "config", "./config/config.yaml", `--config="path_to_config"`)
}

type CommandLineOptions struct {
	ConfigFilePath string
}

var ParseCommandLineArguments = func() CommandLineOptions {
	flag.Parse()
	return CommandLineOptions{ConfigFilePath: config}
}
