package linearize

import (
	"github.com/resonatehq/durable-promise-test-harness/pkg/simulator"
	"github.com/spf13/cobra"
)

var (
	addr     string
	clients  int = 1
	requests int
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "linearize",
		Short:   "run one or more clients and test for linearizability correctness",
		Example: "harness linearize -a http://0.0.0.0:8001/ -r 1000",
		Run: func(cmd *cobra.Command, args []string) {
			sim := simulator.NewSimulation(&simulator.SimulationConfig{
				Addr:        addr,
				NumClients:  clients,
				NumRequests: requests,
				Mode:        simulator.Linearizability,
			})

			if err := sim.Run(); err != nil {
				panic(err)
			}
		},
	}

	cmd.Flags().StringVarP(&addr, "addr", "a", "http://0.0.0.0:8001/", "address of durable-promise server")
	// just one client for now, till porcupine is added
	// cmd.Flags().IntVarP(&clients, "clients", "c", 1, "number of clients")
	cmd.Flags().IntVarP(&requests, "requests", "r", 1, "number of requests per client")

	return cmd
}
