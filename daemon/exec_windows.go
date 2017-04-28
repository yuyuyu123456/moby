package daemon

import (
	"moby/container"
	"moby/daemon/exec"
	"moby/libcontainerd"
)

func execSetPlatformOpt(c *container.Container, ec *exec.Config, p *libcontainerd.Process) error {
	// Process arguments need to be escaped before sending to OCI.
	p.Args = escapeArgs(p.Args)
	p.User.Username = ec.User
	return nil
}
