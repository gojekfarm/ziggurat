package za

import (
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zb"
	"testing"
	"time"
)

func TestZiggurat_Run(t *testing.T) {
	app := NewApp()
	handler := z.HandlerFunc(func(messageEvent zb.MessageEvent, app z.App) z.ProcessStatus {
		return z.ProcessingSuccess
	})
	go func() {
		time.Sleep(500 * time.Millisecond)
		app.Stop()
	}()
	<-app.Run(handler, []string{}, func(opts *RunOptions) {
		opts.StartCallback = func(a z.App) {
			if !app.IsRunning() {
				t.Errorf("expected app run status to be true")
			}
		}
	})
}
