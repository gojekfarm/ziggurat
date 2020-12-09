package cmdparser

import (
	"flag"
	"github.com/gojekfarm/ziggurat/pkg/zb"
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

var ParseCommandLineArguments = func() zb.CommandLineOptions {
	flag.Parse()
	return zb.CommandLineOptions{ConfigFilePath: config}
}
