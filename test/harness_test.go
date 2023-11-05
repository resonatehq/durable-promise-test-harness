package test

import (
	"os"
	"testing"

	"github.com/resonatehq/durable-promise-test-harness/pkg/simulator"
	"github.com/stretchr/testify/suite"
)

func TestMain(t *testing.T) {
	addr := os.Getenv("DP_SERVER")
	if addr == "" {
		addr = "http://0.0.0.0:8001/"
	}

	config := &simulator.SimulationConfig{
		Addr: addr,
	}

	suite.Run(t, simulator.NewSimulation(config))
}
