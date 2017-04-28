package node

import (
	"moby/opts"
)

type nodeOptions struct {
	annotations
	role         string
	availability string
}

type annotations struct {
	name   string
	labels opts.ListOpts
}

func newNodeOptions() *nodeOptions {
	return &nodeOptions{
		annotations: annotations{
			labels: opts.NewListOpts(nil),
		},
	}
}
