// +build !exclude_graphdriver_aufs,linux

package register

import (
	// register the aufs graphdriver
	_ "moby/daemon/graphdriver/aufs"
)
