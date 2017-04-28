package daemon

import (
	apitypes "moby/api/types"
)

// Cluster is the interface for moby/daemon/cluster.(*Cluster).
type Cluster interface {
	ClusterStatus
	NetworkManager
}

// ClusterStatus interface provides information about the Swarm status of the Cluster
type ClusterStatus interface {
	IsAgent() bool
	IsManager() bool
}

// NetworkManager provides methods to manage networks
type NetworkManager interface {
	GetNetwork(input string) (apitypes.NetworkResource, error)
	GetNetworks() ([]apitypes.NetworkResource, error)
	RemoveNetwork(input string) error
}
