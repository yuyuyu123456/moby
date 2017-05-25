/**
 * Created by zizhi.yuwenqi on 2017/5/24.
 */

package filecache
import (
	"github.com/spf13/cobra"

	"github.com/docker/docker/cli"
	"github.com/docker/docker/cli/command"
)

// NewFilecacheCommand returns a cobra command for `filecache` subcommands
func NewFilecacheCommand(dockerCli *command.DockerCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "filecache",
		Short: "Manage filecaches",
		Args:  cli.NoArgs,
		RunE:  dockerCli.ShowHelp,
	}
	cmd.AddCommand(

		newListCommand(dockerCli),
		newRemoveCommand(dockerCli),

	)
	return cmd
}

