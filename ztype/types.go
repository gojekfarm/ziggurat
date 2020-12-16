package ztype

import (
	"github.com/gojekfarm/ziggurat/zbase"
)

type HandlerFunc func(messageEvent zbase.MessageEvent, app App) ProcessStatus

func (h HandlerFunc) HandleMessage(event zbase.MessageEvent, app App) ProcessStatus {
	return h(event, app)
}

type StartFunction func(a App)
type StopFunction func(a App)

const ProcessingSuccess ProcessStatus = 0
const RetryMessage ProcessStatus = 1
const SkipMessage ProcessStatus = 2

type ProcessStatus int
