package ziggurat

type ZigOptions func(z *Ziggurat)

func WithLogger(l StructuredLogger) ZigOptions {
	return func(z *Ziggurat) {
		z.Logger = l
	}
}
