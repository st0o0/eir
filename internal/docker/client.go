package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Client abstracts the Docker SDK calls needed by eir.
type Client interface {
	ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)
	ContainerRestart(ctx context.Context, containerID string, options container.StopOptions) error
	Events(ctx context.Context, options events.ListOptions) (<-chan events.Message, <-chan error)
	Ping(ctx context.Context) (types.Ping, error)
	Close() error
}

// EventFilter builds a Docker event filter for the given container names.
func EventFilter(masterNames []string) filters.Args {
	f := filters.NewArgs()
	f.Add("type", "container")
	f.Add("event", "die")
	f.Add("event", "start")
	for _, name := range masterNames {
		f.Add("container", name)
	}
	return f
}

