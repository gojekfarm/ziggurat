package cmdparser

import (
	"flag"
	"github.com/gojekfarm/ziggurat/zbase"
)

var (
	config string
)

func init() {
	flag.StringVar(&config, "config", "./config/config.yaml", `--config="path_to_config"`)
}

type CommandLineOptions struct {
	ConfigFilePath string
}

var ParseCommandLineArguments = func() zbase.CommandLineOptions {
	flag.Parse()
	return zbase.CommandLineOptions{ConfigFilePath: config}
}
