package pointy

func GetPointer[T comparable](value T) *T {
	return &value
}
