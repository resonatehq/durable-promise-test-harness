package verify

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
		Use:     "verify",
		Short:   "Run multiple concurrent clients to verify for linearizable consistency and performance",
		Example: "harness verify -a http://0.0.0.0:8001/ -r 1000 -c 10",
		Run: func(cmd *cobra.Command, args []string) {
			sim := simulator.NewSimulation(&simulator.SimulationConfig{
				Addr:        addr,
				NumClients:  clients,
				NumRequests: requests,
			})

			if err := sim.Run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Flags().StringVarP(&addr, "addr", "a", "http://0.0.0.0:8001/", "address of durable promise server")
	cmd.Flags().IntVarP(&clients, "clients", "c", 1, "number of clients")
	cmd.Flags().IntVarP(&requests, "requests", "r", 1, "number of requests per client")

	return cmd
}
