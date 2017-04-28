package secret

import (
	"fmt"
	"io"
	"io/ioutil"

	"moby/api/types/swarm"
	"moby/cli"
	"moby/cli/command"
	"moby/opts"
	"moby/pkg/system"
	runconfigopts "moby/runconfig/opts"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type createOptions struct {
	name   string
	file   string
	labels opts.ListOpts
}

func newSecretCreateCommand(dockerCli command.Cli) *cobra.Command {
	createOpts := createOptions{
		labels: opts.NewListOpts(opts.ValidateEnv),
	}

	cmd := &cobra.Command{
		Use:   "create [OPTIONS] SECRET file|-",
		Short: "Create a secret from a file or STDIN as content",
		Args:  cli.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			createOpts.name = args[0]
			createOpts.file = args[1]
			return runSecretCreate(dockerCli, createOpts)
		},
	}
	flags := cmd.Flags()
	flags.VarP(&createOpts.labels, "label", "l", "Secret labels")

	return cmd
}

func runSecretCreate(dockerCli command.Cli, options createOptions) error {
	client := dockerCli.Client()
	ctx := context.Background()

	var in io.Reader = dockerCli.In()
	if options.file != "-" {
		file, err := system.OpenSequential(options.file)
		if err != nil {
			return err
		}
		in = file
		defer file.Close()
	}

	secretData, err := ioutil.ReadAll(in)
	if err != nil {
		return errors.Errorf("Error reading content from %q: %v", options.file, err)
	}

	spec := swarm.SecretSpec{
		Annotations: swarm.Annotations{
			Name:   options.name,
			Labels: runconfigopts.ConvertKVStringsToMap(options.labels.GetAll()),
		},
		Data: secretData,
	}

	r, err := client.SecretCreate(ctx, spec)
	if err != nil {
		return err
	}

	fmt.Fprintln(dockerCli.Out(), r.ID)
	return nil
}
