package ziggurat

import (
	"os"
	"testing"
)

func TestParseCommandLineArgumentsWithDefaultValues(t *testing.T) {
	expected := CommandLineOptions{ConfigFilePath: "./config/config.yaml"}
	cmdOptions := parseCommandLineArguments()
	if expected != cmdOptions {
		t.Errorf("EXPECTED %+v GOT %+v", expected, cmdOptions)
	}

}

func TestParseCommandLineArguments(t *testing.T) {
	os.Args = append(os.Args, "--ziggurat-config=overriddenPath")
	cmdOptions := parseCommandLineArguments()
	newOptions := CommandLineOptions{ConfigFilePath: "overriddenPath"}
	if newOptions != cmdOptions {
		t.Errorf("FAILED got %+v EXPECTED %+v", cmdOptions, newOptions)
	}
}
