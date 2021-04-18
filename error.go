package ziggurat

type ErrStatus string

func (e ErrStatus) Error() string {
	return string(e)
}

const Retry = ErrStatus("retry")
