package ziggurat

type HandlerFunc func(messageEvent MessageEvent, z *Ziggurat) ProcessStatus

func (h HandlerFunc) HandleMessage(event MessageEvent, z *Ziggurat) ProcessStatus {
	return h(event, z)
}

type StartFunction func(z *Ziggurat)
type StopFunction func(z *Ziggurat)

const ProcessingSuccess ProcessStatus = 0
const RetryMessage ProcessStatus = 1
const SkipMessage ProcessStatus = 2

type ProcessStatus int
