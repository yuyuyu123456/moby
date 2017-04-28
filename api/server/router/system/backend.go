package system

import (
	"time"

	"moby/api/types"
	"moby/api/types/events"
	"moby/api/types/filters"
	"golang.org/x/net/context"
)

// Backend is the methods that need to be implemented to provide
// system specific functionality.
type Backend interface {
	SystemInfo() (*types.Info, error)
	SystemVersion() types.Version
	SystemDiskUsage() (*types.DiskUsage, error)
	SubscribeToEvents(since, until time.Time, ef filters.Args) ([]events.Message, chan interface{})
	UnsubscribeFromEvents(chan interface{})
	AuthenticateToRegistry(ctx context.Context, authConfig *types.AuthConfig) (string, string, error)
}
