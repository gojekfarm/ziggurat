package ziggurat

type ZigOptions func(z *Ziggurat)

func WithLogLevel(level string) ZigOptions {
	return func(z *Ziggurat) {
		z.logLevel = level
	}
}
