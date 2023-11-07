package multiple

import (
	"log"

	"github.com/resonatehq/durable-promise-test-harness/pkg/simulator"
	"github.com/spf13/cobra"
)

var (
	addr     string
	clients  int
	requests int
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "multiple",
		Short:   "run multiple concurrent clients and test for linearizability",
		Example: "harness multiple -a http://0.0.0.0:8001/ -c 10 -r 1000",
		Run: func(cmd *cobra.Command, args []string) {
			sim := simulator.NewSimulation(&simulator.SimulationConfig{
				Addr:        addr,
				NumClients:  clients,
				NumRequests: requests,
			})

			if err := sim.Multiple(); err != nil {
				panic(err)
			}

			log.Printf("multiple test passed!\n")
		},
	}

	cmd.Flags().StringVarP(&addr, "addr", "a", "http://0.0.0.0:8001/", "address of durable-promise server")
	cmd.Flags().IntVarP(&clients, "clients", "c", 1, "number of clients")
	cmd.Flags().IntVarP(&requests, "requests", "r", 1, "number of requests per client")

	return cmd
}
