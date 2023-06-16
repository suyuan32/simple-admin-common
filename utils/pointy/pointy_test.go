package pointy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPointer(t *testing.T) {
	var dataIntOrigin int = 1
	dataInt := GetPointer(dataIntOrigin)
	assert.Equal(t, dataIntOrigin, *dataInt)

	var dataUintOrigin uint = 1
	dataUint := GetPointer(dataUintOrigin)
	assert.Equal(t, dataUintOrigin, *dataUint)

	var dataStringOrigin = "123"
	dataString := GetPointer(dataStringOrigin)
	assert.Equal(t, dataStringOrigin, *dataString)
}
