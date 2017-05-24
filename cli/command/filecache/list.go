/**
 * Created by zizhi.yuwenqi on 2017/5/22.
 */

package filecache

import (
	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/cli"
	"github.com/docker/docker/cli/command"
	"github.com/docker/docker/cli/command/formatter"
	"github.com/docker/docker/opts"
	"github.com/spf13/cobra"
)


type filecachesOptions struct {
	matchName string
	quiet       bool
	showDigests bool
	format      string
	filter      opts.FilterOpt
}

// NewImagesCommand creates a new `docker images` command
func NewFileCachesCommand(dockerCli *command.DockerCli) *cobra.Command {
	opts := filecachesOptions{filter: opts.NewFilterOpt()}
	cmd := &cobra.Command{
		Use:   "filecaches [OPTIONS] [orig]",
		Short: "List filecaches",
		Args:  cli.RequiresMaxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.matchName = args[0]
			}
			return runFilecaches(dockerCli, opts)
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&opts.quiet, "quiet", "q", false, "Only show filecache orig")
	//flags.BoolVarP(&opts.all, "all", "a", false, "Show all images (default hides intermediate images)")
	//flags.BoolVar(&opts.noTrunc, "no-trunc", false, "Don't truncate output")
	//flags.BoolVar(&opts.showDigests, "digests", false, "Show digests")
	//flags.StringVar(&opts.format, "format", "", "Pretty-print images using a Go template")
	flags.VarP(&opts.filter, "filter", "f", "Filter output based on conditions provided")

	return cmd
}

func newListCommand(dockerCli *command.DockerCli) *cobra.Command {
	cmd := *NewFileCachesCommand(dockerCli)
	cmd.Aliases = []string{"filecaches", "list"}
	cmd.Use = "ls [OPTIONS] [orig]"
	return &cmd
}

func runFilecaches(dockerCli *command.DockerCli, opts filecachesOptions) error {
	ctx := context.Background()

	filters := opts.filter.Value()
	if opts.matchName != "" {
		filters.Add("reference", opts.matchName)
	}

	options := types.FileCachesOptions{
		Filters: filters,
	}

	filecaches, err := dockerCli.Client().FileCacheList(ctx, options)
	if err != nil {
		return err
	}

	format := opts.format
	if len(format) == 0 {
		//if len(dockerCli.ConfigFile().ImagesFormat) > 0 && !opts.quiet {
		//	format = dockerCli.ConfigFile().ImagesFormat
		//} else {
			format = formatter.TableFormatKey
		//}
	}

	//imageCtx := formatter.ImageContext{
	//	Context: formatter.Context{
	//		Output: dockerCli.Out(),
	//		Format: formatter.NewImageFormat(format, opts.quiet, opts.showDigests),
	//		Trunc:  !opts.noTrunc,
	//	},
	//	Digest: opts.showDigests,
	//}
	filecachesCtx:=formatter.FileCacheContext{
		Context: formatter.Context{
			Output: dockerCli.Out(),
			Format: formatter.NewFileCacheFormat(format, opts.quiet, opts.showDigests),
		},
		Digest: opts.showDigests,
	}
	return formatter.FileCacheWrite(filecachesCtx, filecaches)
}

