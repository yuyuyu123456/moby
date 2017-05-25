/**
 * Created by zizhi.yuwenqi on 2017/5/25.
 */

package filecache

import (
	"fmt"
	"strings"
	"github.com/docker/docker/cli/command"
	"github.com/spf13/cobra"
	"github.com/docker/docker/cli"
	"golang.org/x/net/context"
	"github.com/pkg/errors"
	"github.com/Sirupsen/logrus"
)


// NewRemoveCommand creates a new `docker remove` command
func NewRemoveCommand(dockerCli *command.DockerCli) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "rmf FileCache [FileCache...]",
		Short: "Remove one or more filecaches",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemove(dockerCli, args)
		},
	}

	//flags := cmd.Flags()
	return cmd
}

func newRemoveCommand(dockerCli *command.DockerCli) *cobra.Command {
	cmd := *NewRemoveCommand(dockerCli)
	cmd.Aliases = []string{"rmf", "remove"}
	cmd.Use = "rm FileCache [FileCache...]"
	return &cmd
}

func runRemove(dockerCli *command.DockerCli, filecaches []string) error {
	client := dockerCli.Client()
	ctx := context.Background()
        logrus.Debug("runRemove filecaches :",filecaches)
	var errs []string
	for _, filecache := range filecaches{
		dels, err := client.FileCacheRemove(ctx,filecache)
		if err != nil {
			logrus.Debug("runRemove error:",err)
			errs = append(errs, err.Error())
		} else {
			for _, del := range dels {
				if del.Orig != "" {
					fmt.Fprintf(dockerCli.Out(), "Deleted: %s\n", del.Orig)
				} else {
					fmt.Fprintf(dockerCli.Out(), " %s\n", del.Notexist)
				}
			}
		}
	}

	if len(errs) > 0 {
		return errors.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}

