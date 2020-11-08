package zig

import "flag"

var (
	config string
)

func init() {
	flag.StringVar(&config, "appconf", "./appconf/appconf.yaml", `--appconf="path_to_config"`)
}

type CommandLineOptions struct {
	ConfigFilePath string
}

func ParseCommandLineArguments() CommandLineOptions {
	flag.Parse()
	return CommandLineOptions{ConfigFilePath: config}
}
