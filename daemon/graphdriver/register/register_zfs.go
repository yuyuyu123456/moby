// +build !exclude_graphdriver_zfs,linux !exclude_graphdriver_zfs,freebsd, solaris

package register

import (
	// register the zfs driver
	_ "moby/daemon/graphdriver/zfs"
)
