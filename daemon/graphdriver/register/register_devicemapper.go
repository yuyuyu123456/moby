// +build !exclude_graphdriver_devicemapper,linux

package register

import (
	// register the devmapper graphdriver
	_ "moby/daemon/graphdriver/devmapper"
)
