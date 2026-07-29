package detector

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"

	dockerclient "github.com/st0o0/eir/internal/docker"
	"github.com/st0o0/eir/internal/watcher"
)

// Classification describes why a master container reappeared.
type Classification int

const (
	RestartCase  Classification = iota
	RecreateCase
)

func (c Classification) String() string {
	switch c {
	case RestartCase:
		return "restart"
	case RecreateCase:
		return "recreate"
	default:
		return "unknown"
	}
}

// Dependent represents a container that depends on a master's network.
type Dependent struct {
	ContainerID string
	Name        string
	WasRunning  bool
}

// Detector classifies master events and discovers dependent containers.
type Detector struct {
	client        dockerclient.Client
	discoveryMode string
	logger        *slog.Logger
	lastKnownIDs  map[string]string
}

// New creates a Detector with the given discovery mode ("inspect" or "label").
func New(client dockerclient.Client, discoveryMode string, logger *slog.Logger) *Detector {
	return &Detector{
		client:        client,
		discoveryMode: discoveryMode,
		logger:        logger,
		lastKnownIDs:  make(map[string]string),
	}
}

// Classify processes an event and returns the classification and affected dependents.
// Returns false for the ok value when the event should not trigger healing (e.g. die events).
func (d *Detector) Classify(ctx context.Context, event watcher.Event) (Classification, []Dependent, bool) {
	if event.Action == "die" {
		d.logger.Info("master died, waiting for restart",
			"master", event.MasterName,
			"container_id", event.ContainerID[:12],
		)
		return 0, nil, false
	}

	lastID, known := d.lastKnownIDs[event.MasterName]
	d.lastKnownIDs[event.MasterName] = event.ContainerID

	var classification Classification
	if known && lastID == event.ContainerID {
		classification = RestartCase
	} else {
		classification = RecreateCase
	}

	d.logger.Info("classified event",
		"master", event.MasterName,
		"classification", classification.String(),
		"container_id", event.ContainerID[:12],
	)

	dependents, err := d.discoverDependents(ctx, event.MasterName, event.ContainerID)
	if err != nil {
		d.logger.Error("failed to discover dependents",
			"master", event.MasterName,
			"error", err,
		)
		return 0, nil, false
	}

	return classification, dependents, true
}

func (d *Detector) discoverDependents(ctx context.Context, masterName, masterID string) ([]Dependent, error) {
	switch d.discoveryMode {
	case "label":
		return d.discoverByLabel(ctx, masterName)
	default:
		return d.discoverByInspect(ctx, masterName, masterID)
	}
}

func (d *Detector) discoverByInspect(ctx context.Context, masterName, masterID string) ([]Dependent, error) {
	containers, err := d.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}

	var dependents []Dependent
	for _, c := range containers {
		if c.ID == masterID {
			continue
		}

		info, err := d.client.ContainerInspect(ctx, c.ID)
		if err != nil {
			d.logger.Warn("failed to inspect container",
				"container_id", c.ID[:12],
				"error", err,
			)
			continue
		}

		networkMode := string(info.HostConfig.NetworkMode)
		// network_mode: container:<name_or_id>
		if strings.HasPrefix(networkMode, "container:") {
			target := strings.TrimPrefix(networkMode, "container:")
			if target == masterName || target == masterID {
				name := strings.TrimPrefix(info.Name, "/")
				dependents = append(dependents, Dependent{
					ContainerID: c.ID,
					Name:        name,
					WasRunning:  info.State.Running,
				})
			}
		}
	}

	return dependents, nil
}

func (d *Detector) discoverByLabel(ctx context.Context, masterName string) ([]Dependent, error) {
	f := filters.NewArgs()
	f.Add("label", fmt.Sprintf("eir.master=%s", masterName))

	containers, err := d.client.ContainerList(ctx, container.ListOptions{All: true, Filters: f})
	if err != nil {
		return nil, fmt.Errorf("listing labeled containers: %w", err)
	}

	var dependents []Dependent
	for _, c := range containers {
		info, err := d.client.ContainerInspect(ctx, c.ID)
		if err != nil {
			d.logger.Warn("failed to inspect container",
				"container_id", c.ID[:12],
				"error", err,
			)
			continue
		}

		name := strings.TrimPrefix(info.Name, "/")
		dependents = append(dependents, Dependent{
			ContainerID: c.ID,
			Name:        name,
			WasRunning:  info.State.Running,
		})
	}

	return dependents, nil
}
