package executor

import (
	"io"
	"time"

	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	"moby/api/types"
	"moby/api/types/backend"
	"moby/api/types/container"
	"moby/api/types/events"
	"moby/api/types/filters"
	"moby/api/types/network"
	swarmtypes "moby/api/types/swarm"
	clustertypes "moby/daemon/cluster/provider"
	"moby/plugin"
	"github.com/docker/libnetwork"
	"github.com/docker/libnetwork/cluster"
	networktypes "github.com/docker/libnetwork/types"
	"github.com/docker/swarmkit/agent/exec"
	"golang.org/x/net/context"
)

// Backend defines the executor component for a swarm agent.
type Backend interface {
	CreateManagedNetwork(clustertypes.NetworkCreateRequest) error
	DeleteManagedNetwork(name string) error
	FindNetwork(idName string) (libnetwork.Network, error)
	SetupIngress(clustertypes.NetworkCreateRequest, string) (<-chan struct{}, error)
	ReleaseIngress() (<-chan struct{}, error)
	PullImage(ctx context.Context, image, tag string, metaHeaders map[string][]string, authConfig *types.AuthConfig, outStream io.Writer) error
	CreateManagedContainer(config types.ContainerCreateConfig) (container.ContainerCreateCreatedBody, error)
	ContainerStart(name string, hostConfig *container.HostConfig, checkpoint string, checkpointDir string) error
	ContainerStop(name string, seconds *int) error
	ContainerLogs(context.Context, string, *types.ContainerLogsOptions) (<-chan *backend.LogMessage, error)
	ConnectContainerToNetwork(containerName, networkName string, endpointConfig *network.EndpointSettings) error
	ActivateContainerServiceBinding(containerName string) error
	DeactivateContainerServiceBinding(containerName string) error
	UpdateContainerServiceConfig(containerName string, serviceConfig *clustertypes.ServiceConfig) error
	ContainerInspectCurrent(name string, size bool) (*types.ContainerJSON, error)
	ContainerWaitWithContext(ctx context.Context, name string) error
	ContainerRm(name string, config *types.ContainerRmConfig) error
	ContainerKill(name string, sig uint64) error
	SetContainerSecretStore(name string, store exec.SecretGetter) error
	SetContainerSecretReferences(name string, refs []*swarmtypes.SecretReference) error
	SystemInfo() (*types.Info, error)
	VolumeCreate(name, driverName string, opts, labels map[string]string) (*types.Volume, error)
	Containers(config *types.ContainerListOptions) ([]*types.Container, error)
	SetNetworkBootstrapKeys([]*networktypes.EncryptionKey) error
	DaemonJoinsCluster(provider cluster.Provider)
	DaemonLeavesCluster()
	IsSwarmCompatible() error
	SubscribeToEvents(since, until time.Time, filter filters.Args) ([]events.Message, chan interface{})
	UnsubscribeFromEvents(listener chan interface{})
	UpdateAttachment(string, string, string, *network.NetworkingConfig) error
	WaitForDetachment(context.Context, string, string, string, string) error
	GetRepository(context.Context, reference.NamedTagged, *types.AuthConfig) (distribution.Repository, bool, error)
	LookupImage(name string) (*types.ImageInspect, error)
	PluginManager() *plugin.Manager
	PluginGetter() *plugin.Store
}
