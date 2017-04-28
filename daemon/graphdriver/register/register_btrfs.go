// +build !exclude_graphdriver_btrfs,linux

package register

import (
	// register the btrfs graphdriver
	_ "moby/daemon/graphdriver/btrfs"
)
