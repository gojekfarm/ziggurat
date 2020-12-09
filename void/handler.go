package void

import (
	"github.com/gojekfarm/ziggurat/z"
	"github.com/gojekfarm/ziggurat/zb"
)

type VoidMessageHandler struct{}

func (v VoidMessageHandler) HandleMessage(event zb.MessageEvent, app z.App) z.ProcessStatus {
	return z.ProcessingSuccess
}
