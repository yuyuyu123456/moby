//+build windows

package daemon

import (
	"moby/container"
)

func (daemon *Daemon) saveApparmorConfig(container *container.Container) error {
	return nil
}
