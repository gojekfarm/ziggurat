package kafka

import (
	"errors"
)

var (
	ErrOffsetCommit = errors.New("cannot commit errored message")
)
