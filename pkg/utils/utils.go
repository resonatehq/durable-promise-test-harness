package utils

import (
	"os"
	"path"

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

func WriteStringToFile(content string, filepath string) error {
	data := []byte(content)
	err := os.MkdirAll(path.Dir(filepath), 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}
