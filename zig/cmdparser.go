package zig

import "flag"

var (
	config string
)

func init() {
	flag.StringVar(&config, "ziggurat-config", "./config/config.yaml", `--ziggurat-config="path_to_config"`)
}

type CommandLineOptions struct {
	ConfigFilePath string
}

func parseCommandLineArguments() CommandLineOptions {
	flag.Parse()
	return CommandLineOptions{ConfigFilePath: config}
}
