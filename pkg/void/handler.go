package void

import (
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
)

type VoidMessageHandler struct{}

func (v VoidMessageHandler) HandleMessage(event basic.MessageEvent, app z.App) z.ProcessStatus {
	return z.ProcessingSuccess
}
