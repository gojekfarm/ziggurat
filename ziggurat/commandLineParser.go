package ziggurat

import "flag"

type CommandLineOptions struct {
	ConfigFilePath string
}

func parseCommandLineArguments() CommandLineOptions {
	defaultConfigPath := "./config/config.yaml"
	configPath := ""
	flag.StringVar(&configPath, defaultConfigPath, defaultConfigPath, `--config="path_to_config_file"`)
	flag.Parse()
	return CommandLineOptions{ConfigFilePath: configPath}
}
