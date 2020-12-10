package cmdparser_test

import (
	"github.com/gojekfarm/ziggurat/cmdparser"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/zlog"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	zlog.ConfigureLogger("disabled")
	os.Exit(m.Run())
}

func TestParseCommandLineArgumentsWithDefaultValues(t *testing.T) {
	expected := zbase.CommandLineOptions{ConfigFilePath: "./config/config.yaml"}
	cmdOptions := cmdparser.ParseCommandLineArguments()
	if expected != cmdOptions {
		t.Errorf("EXPECTED %+v GOT %+v", expected, cmdOptions)
	}

}

func TestParseCommandLineArguments(t *testing.T) {
	os.Args = append(os.Args, "--config=overriddenPath")
	cmdOptions := cmdparser.ParseCommandLineArguments()
	newOptions := zbase.CommandLineOptions{ConfigFilePath: "overriddenPath"}
	if newOptions != cmdOptions {
		t.Errorf("FAILED got %+v EXPECTED %+v", cmdOptions, newOptions)
	}
}
