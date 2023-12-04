// Copyright 2023 The Ryan SU Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pointy

import "time"

var zeroTime time.Time

// GetPointer returns pointer from comparable type
func GetPointer[T comparable](value T) *T {
	return &value
}

// GetSlicePointer returns pointer from comparable type slice
func GetSlicePointer[T comparable](value []T) (result []*T) {
	for _, v := range value {
		result = append(result, GetPointer(v))
	}
	return result
}

// GetStatusPointer returns the status pointer in rpc
func GetStatusPointer(value *uint32) (result *uint8) {
	if value == nil {
		return nil
	}

	result = new(uint8)

	*result = uint8(*value)

	return result
}

// GetTimePointer returns the time pointer from int64
func GetTimePointer(value *int64, nsec int64) (result *time.Time) {
	if value == nil {
		return nil
	}

	result = new(time.Time)

	*result = time.Unix(*value, nsec)

	return result
}

// GetUnixMilliPointer returns nil when int64 is -621355968000, your time should not be from before 1970
//
// Example:
//
//	 var zeroTime time.Time
//		zeroTimeP := GetUnixMilliPointer(zeroTime)
//		fmt.Println(zeroTimeP)
//
// Result:  <nil>
func GetUnixMilliPointer(value int64) *int64 {
	if value == zeroTime.UnixMilli() {
		return nil
	}
	return &value
}
