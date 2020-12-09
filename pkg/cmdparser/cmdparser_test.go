package cmdparser_test

import (
	"github.com/gojekfarm/ziggurat/pkg/cmdparser"
	"github.com/gojekfarm/ziggurat/pkg/zb"
	"github.com/gojekfarm/ziggurat/pkg/zlog"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	zlog.ConfigureLogger("disabled")
	os.Exit(m.Run())
}

func TestParseCommandLineArgumentsWithDefaultValues(t *testing.T) {
	expected := zb.CommandLineOptions{ConfigFilePath: "./config/config.yaml"}
	cmdOptions := cmdparser.ParseCommandLineArguments()
	if expected != cmdOptions {
		t.Errorf("EXPECTED %+v GOT %+v", expected, cmdOptions)
	}

}

func TestParseCommandLineArguments(t *testing.T) {
	os.Args = append(os.Args, "--config=overriddenPath")
	cmdOptions := cmdparser.ParseCommandLineArguments()
	newOptions := zb.CommandLineOptions{ConfigFilePath: "overriddenPath"}
	if newOptions != cmdOptions {
		t.Errorf("FAILED got %+v EXPECTED %+v", cmdOptions, newOptions)
	}
}
