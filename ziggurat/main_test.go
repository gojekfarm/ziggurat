package ziggurat

import (
	"github.com/rs/zerolog"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	turnOffLogger()
	os.Exit(m.Run())
}

func turnOffLogger() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}
