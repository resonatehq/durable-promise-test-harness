package load

import (
	"log"

	"github.com/resonatehq/durable-promise-test-harness/pkg/simulator"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "load",
		Short:   "",
		Example: "",
		Run: func(cmd *cobra.Command, args []string) {
			sim := simulator.NewSimulation(&simulator.SimulationConfig{})

			if err := sim.Load(); err != nil {
				panic(err)
			}

			log.Printf("load test passed!\n")
		},
	}

	return cmd
}
