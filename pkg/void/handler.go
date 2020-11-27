package void

import (
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zbasic"
)

type VoidMessageHandler struct{}

func (v VoidMessageHandler) HandleMessage(event zbasic.MessageEvent, app z.App) z.ProcessStatus {
	return z.ProcessingSuccess
}
