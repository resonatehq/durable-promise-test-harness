package cmd

import (
	"github.com/resonatehq/durable-promise-test-harness/cmd/verify"
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
			Message: "Load test commands",
			Commands: []*cobra.Command{
				verify.NewCmd(),
			},
		},
	}

	groups.Add(rootCmd)

	return rootCmd
}
