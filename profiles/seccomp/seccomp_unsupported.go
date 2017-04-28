// +build linux,!seccomp

package seccomp

import (
	"moby/api/types"
	"github.com/opencontainers/runtime-spec/specs-go"
)

// DefaultProfile returns a nil pointer on unsupported systems.
func DefaultProfile() *types.Seccomp {
	return nil
}
