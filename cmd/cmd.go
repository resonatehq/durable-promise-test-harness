package cmd

import (
	"github.com/resonatehq/durable-promise-test-harness/cmd/load"
	"github.com/resonatehq/durable-promise-test-harness/cmd/multiple"
	"github.com/resonatehq/durable-promise-test-harness/cmd/single"
	"github.com/resonatehq/durable-promise-test-harness/pkg/utils"
	"github.com/spf13/cobra"
)

func NewDPCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "harness",
		Short: "durable promise server testing harness",
	}

	groups := utils.CommandGroups{
		{
			Message: "Single client correctness test commands",
			Commands: []*cobra.Command{
				single.NewCmd(),
			},
		},
		{
			Message: "Multiple client linearizability test commands",
			Commands: []*cobra.Command{
				multiple.NewCmd(),
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
