package ziggurat

type HandlerFunc func(event Event) ProcessStatus

func (h HandlerFunc) HandleEvent(event Event) ProcessStatus {
	return h(event)
}
