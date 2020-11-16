package cmdparser

import (
	"flag"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
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

var ParseCommandLineArguments = func() basic.CommandLineOptions {
	flag.Parse()
	return basic.CommandLineOptions{ConfigFilePath: config}
}
