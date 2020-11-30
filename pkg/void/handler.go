package void

import (
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zb"
)

type VoidMessageHandler struct{}

func (v VoidMessageHandler) HandleMessage(event zb.MessageEvent, app z.App) z.ProcessStatus {
	return z.ProcessingSuccess
}
