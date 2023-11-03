package utils

import (
	"github.com/mohae/deepcopy"
)

func ToPointer[T any](val T) *T {
	v := val
	return &v
}

func DeepCopy[T any](input T) T {
	return deepcopy.Copy(input).(T)
}
