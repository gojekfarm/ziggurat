package ziggurat

import (
	"context"
	"github.com/sethvargo/go-signalcontext"
	"syscall"
)

type ZigOptions func(z *Ziggurat)

func WithLogLevel(level string) ZigOptions {
	return func(z *Ziggurat) {
		z.logLevel = level
	}
}

func WithContext(ctx context.Context) ZigOptions {
	return func(z *Ziggurat) {
		newCtx, canceler := signalcontext.Wrap(ctx, syscall.SIGABRT, syscall.SIGINT)
		z.ctx = newCtx
		z.cancelFun = canceler
	}
}
