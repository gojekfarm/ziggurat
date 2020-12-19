package ziggurat

type ZigOptions func(z *Ziggurat)

func WithLogger(l LeveledLogger) ZigOptions {
	return func(z *Ziggurat) {
		z.Logger = l
	}
}
