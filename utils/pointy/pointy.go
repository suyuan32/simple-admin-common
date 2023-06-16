package pointy

import "time"

// GetPointer returns pointer from comparable type
func GetPointer[T comparable](value T) *T {
	return &value
}

// GetStatusPointer returns the status pointer in rpc
func GetStatusPointer(value *uint32) *uint8 {
	if value == nil {
		return nil
	}

	var result *uint8

	*result = uint8(*value)

	return result
}

// GetTimePointer returns the time pointer from int64
func GetTimePointer(value *int64, nsec int64) *time.Time {
	if value == nil {
		return nil
	}

	var result *time.Time

	*result = time.Unix(*value, nsec)

	return result
}
