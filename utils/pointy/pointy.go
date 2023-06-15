package pointy

// GetPointer returns pointer from comparable type
func GetPointer[T comparable](value T) *T {
	return &value
}
