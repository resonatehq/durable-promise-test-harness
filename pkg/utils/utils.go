package utils

import (
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/spf13/cobra"
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

type CommandGroup struct {
	Message  string
	Commands []*cobra.Command
}

type CommandGroups []CommandGroup

func (g CommandGroups) Add(c *cobra.Command) {
	for _, group := range g {
		c.AddCommand(group.Commands...)
	}
}

func IsReady(Addr string) bool {
	serverAddr := strings.TrimSuffix(strings.TrimPrefix(Addr, "http://"), "/")
	conn, err := net.DialTimeout("tcp", serverAddr, 1*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
