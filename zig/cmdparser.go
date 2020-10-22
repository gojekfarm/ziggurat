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

func ParseCommandLineArguments() CommandLineOptions {
	flag.Parse()
	return CommandLineOptions{ConfigFilePath: config}
}
