package ziggurat

type ErrProcessingFailed struct {
	Action string
}

func (e ErrProcessingFailed) Error() string {
	return "message processing failed"
}
