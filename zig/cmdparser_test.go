package zig

import (
	"os"
	"testing"
)

func TestParseCommandLineArgumentsWithDefaultValues(t *testing.T) {
	expected := CommandLineOptions{ConfigFilePath: "./appconf/appconf.yaml"}
	cmdOptions := ParseCommandLineArguments()
	if expected != cmdOptions {
		t.Errorf("EXPECTED %+v GOT %+v", expected, cmdOptions)
	}

}

func TestParseCommandLineArguments(t *testing.T) {
	os.Args = append(os.Args, "--appconf=overriddenPath")
	cmdOptions := ParseCommandLineArguments()
	newOptions := CommandLineOptions{ConfigFilePath: "overriddenPath"}
	if newOptions != cmdOptions {
		t.Errorf("FAILED got %+v EXPECTED %+v", cmdOptions, newOptions)
	}
}
