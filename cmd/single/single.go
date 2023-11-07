package single

import (
	"log"

	"github.com/resonatehq/durable-promise-test-harness/pkg/simulator"
	"github.com/spf13/cobra"
)

var (
	addr     string
	requests int
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "single",
		Short:   "",
		Example: "",
		Run: func(cmd *cobra.Command, args []string) {
			sim := simulator.NewSimulation(&simulator.SimulationConfig{
				Addr:        addr,
				NumClients:  1,
				NumRequests: requests,
			})

			if err := sim.Single(); err != nil {
				panic(err)
			}

			log.Printf("single client correctness validation passed!\n")
		},
	}

	cmd.Flags().StringVarP(&addr, "addr", "a", "http://0.0.0.0:8001/", "address of durable-promise server")
	cmd.Flags().IntVarP(&requests, "requests", "r", 1, "number of requests per client")

	return cmd
}
