package system

import (
	"github.com/spf13/cobra"

	"moby/cli"
	"moby/cli/command"
)

// NewSystemCommand returns a cobra command for `system` subcommands
func NewSystemCommand(dockerCli *command.DockerCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "Manage Docker",
		Args:  cli.NoArgs,
		RunE:  dockerCli.ShowHelp,
	}
	cmd.AddCommand(
		NewEventsCommand(dockerCli),
		NewInfoCommand(dockerCli),
		NewDiskUsageCommand(dockerCli),
		NewPruneCommand(dockerCli),
	)

	return cmd
}
