package test

import (
	"os"
	"testing"

	"github.com/resonatehq/durable-promise-test-harness/pkg/simulator"
	"github.com/stretchr/testify/suite"
)

// SERVER_ADDR=http://0.0.0.0:8001/ go test -v ./test/...
func TestMain(t *testing.T) {
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = "http://0.0.0.0:8001/"
	}

	config := &simulator.SimulationConfig{
		Addr: addr,
	}

	suite.Run(t, simulator.New(config))
}
