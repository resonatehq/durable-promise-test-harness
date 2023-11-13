package verify

import (
	"log"

	"github.com/resonatehq/durable-promise-test-harness/pkg/simulator"
	sim_sub "github.com/resonatehq/durable-promise-test-harness/pkg/simulator/subscriptions"
	"github.com/spf13/cobra"
)

var (
	addr          string
	clients       int
	requests      int
	subscriptions bool
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "verify",
		Short:   "Run multiple concurrent clients to verify for linearizable consistency and performance",
		Example: "harness verify -a http://0.0.0.0:8001/ -r 1000 -c 10",
		Run: func(cmd *cobra.Command, args []string) {
			if subscriptions {
				sim_sub.Run()
				return
			}

			sim := simulator.NewSimulation(&simulator.SimulationConfig{
				Addr:        addr,
				NumClients:  clients,
				NumRequests: requests,
				Mode:        simulator.Load,
			})

			if err := sim.Run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Flags().StringVarP(&addr, "addr", "a", "http://0.0.0.0:8001/", "address of durable promise server")
	cmd.Flags().IntVarP(&clients, "clients", "c", 1, "number of clients")
	cmd.Flags().IntVarP(&requests, "requests", "r", 1, "number of requests per client")

	cmd.Flags().BoolVarP(&subscriptions, "test-subscription", "s", false, "run experimental subscriptions api testing")
	cmd.Flags().MarkHidden("test-subscription")

	return cmd
}
