package cmd

import (
	"github.com/resonatehq/durable-promise-test-harness/cmd/linearize"
	"github.com/resonatehq/durable-promise-test-harness/cmd/load"
	"github.com/resonatehq/durable-promise-test-harness/pkg/utils"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "harness",
		Short: "durable promise server testing harness",
	}

	groups := utils.CommandGroups{
		{
			Message: "Linearizability test commands",
			Commands: []*cobra.Command{
				linearize.NewCmd(),
			},
		},
		{
			Message: "Load test commands",
			Commands: []*cobra.Command{
				load.NewCmd(),
			},
		},
	}

	groups.Add(rootCmd)

	return rootCmd
}
