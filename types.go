package ziggurat

type HandlerFunc func(messageEvent Event, app AppContext) ProcessStatus

func (h HandlerFunc) HandleMessage(event Event, app AppContext) ProcessStatus {
	return h(event, app)
}

type StartFunction func(a AppContext)
type StopFunction func(a AppContext)

const ProcessingSuccess ProcessStatus = 0
const RetryMessage ProcessStatus = 1
const SkipMessage ProcessStatus = 2

type ProcessStatus int
