package cmdparser

import (
	"github.com/gojekfarm/ziggurat-go/pkg/ziggurat/basic"
	"os"
	"testing"
)

func TestParseCommandLineArgumentsWithDefaultValues(t *testing.T) {
	expected := basic.CommandLineOptions{ConfigFilePath: "./config/config.yaml"}
	cmdOptions := ParseCommandLineArguments()
	if expected != cmdOptions {
		t.Errorf("EXPECTED %+v GOT %+v", expected, cmdOptions)
	}

}

func TestParseCommandLineArguments(t *testing.T) {
	os.Args = append(os.Args, "--config=overriddenPath")
	cmdOptions := ParseCommandLineArguments()
	newOptions := basic.CommandLineOptions{ConfigFilePath: "overriddenPath"}
	if newOptions != cmdOptions {
		t.Errorf("FAILED got %+v EXPECTED %+v", cmdOptions, newOptions)
	}
}
