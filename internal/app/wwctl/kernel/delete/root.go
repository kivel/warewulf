package delete

import (
	"github.com/hpcng/warewulf/internal/pkg/kernel"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "delete [flags] [kernel version]...",
		Short: "Delete an imported kernel",
		Long:  "This command will delete a kernel that has been imported into Warewulf.",
		RunE:  CobraRunE,
		Args:  cobra.MinimumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			list, _ := kernel.ListKernels()
			return list, cobra.ShellCompDirectiveNoFileComp
		},
	}
)

func init() {

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
