package utils

import (
	"github.com/mohae/deepcopy"
)

func ToPointer[T any](val T) *T {
	v := val
	return &v
}

func DeepCopy[T any](val T) T {
	return deepcopy.Copy(val).(T)
}

func SafeDereference[T any](val *T) T {
	if val == nil {
		var t T
		return t
	}
	return *val
}
