package zvoid

import (
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/ztype"
)

type VoidMessageHandler struct{}

func (v VoidMessageHandler) HandleMessage(event zbase.MessageEvent, app ztype.App) ztype.ProcessStatus {
	return ztype.ProcessingSuccess
}
