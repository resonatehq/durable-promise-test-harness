package main

import (
	"testing"

	"github.com/resonatehq/durable-promise-test-harness/pkg/simulator"
)

// The Go coverage tool only works in conjunction with the testing package
func TestHarness(t *testing.T) {
	t.Run("TestHarness", func(t *testing.T) {
		sim := simulator.NewSimulation(&simulator.SimulationConfig{
			Addr:        "http://0.0.0.0:8001/",
			NumClients:  10,
			NumRequests: 1,
		})

		t.Log("Starting")

		if err := sim.Run(); err != nil {
			t.Fatal(err)
		}

		t.Log("Finished")
	})
}
