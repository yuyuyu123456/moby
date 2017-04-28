// +build windows

package daemon

import (
	"moby/api/types/container"
	"moby/libcontainerd"
)

func toContainerdResources(resources container.Resources) libcontainerd.Resources {
	var r libcontainerd.Resources
	return r
}
