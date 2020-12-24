package ziggurat

type HandlerFunc func(event Event) ProcessStatus

func (h HandlerFunc) HandleMessage(event Event) ProcessStatus {
	return h(event)
}
