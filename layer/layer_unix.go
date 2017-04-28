// +build linux freebsd darwin openbsd solaris

package layer

import "moby/pkg/stringid"

func (ls *layerStore) mountID(name string) string {
	return stringid.GenerateRandomID()
}
