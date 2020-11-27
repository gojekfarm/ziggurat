package cmdparser

import (
	"flag"
	"github.com/gojekfarm/ziggurat-go/pkg/zbasic"
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

var ParseCommandLineArguments = func() zbasic.CommandLineOptions {
	flag.Parse()
	return zbasic.CommandLineOptions{ConfigFilePath: config}
}
