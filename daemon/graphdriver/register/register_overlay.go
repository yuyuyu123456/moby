// +build !exclude_graphdriver_overlay,linux

package register

import (
	// register the overlay graphdriver
	_ "moby/daemon/graphdriver/overlay"
	_ "moby/daemon/graphdriver/overlay2"
)
