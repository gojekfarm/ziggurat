package ziggurat

type MessageHandler interface {
	HandleMessage(event MessageEvent, z *Ziggurat) ProcessStatus
}

type Streams interface {
	Start(z *Ziggurat) (chan struct{}, error)
	Stop()
}
