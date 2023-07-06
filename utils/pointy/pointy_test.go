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

func TestGetSlicePointer(t *testing.T) {
	nums := []int{1, 2, 3}
	numsPointer := GetSlicePointer(nums)
	assert.Equal(t, nums[0], *numsPointer[0])
	assert.Equal(t, nums[1], *numsPointer[1])
	assert.Equal(t, nums[2], *numsPointer[2])

	strs := []string{"a", "b", "c"}
	strsPointer := GetSlicePointer(strs)
	assert.Equal(t, strs[0], *strsPointer[0])
	assert.Equal(t, strs[1], *strsPointer[1])
	assert.Equal(t, strs[2], *strsPointer[2])
}
