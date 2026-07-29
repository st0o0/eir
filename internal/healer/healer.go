package healer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/docker/docker/api/types/container"

	dockerclient "github.com/st0o0/eir/internal/docker"
	"github.com/st0o0/eir/internal/detector"
)

// Healer executes recovery actions on dependent containers.
type Healer struct {
	client        dockerclient.Client
	stabilizeWait time.Duration
	maxRetries    int
	retryBackoff  time.Duration
	logger        *slog.Logger
}

// New creates a Healer with the given configuration.
func New(client dockerclient.Client, stabilizeWait time.Duration, maxRetries int, retryBackoff time.Duration, logger *slog.Logger) *Healer {
	return &Healer{
		client:        client,
		stabilizeWait: stabilizeWait,
		maxRetries:    maxRetries,
		retryBackoff:  retryBackoff,
		logger:        logger,
	}
}

// Heal restores network connectivity for dependent containers after a master event.
func (h *Healer) Heal(ctx context.Context, classification detector.Classification, masterID string, masterName string, dependents []detector.Dependent) error {
	if len(dependents) == 0 {
		h.logger.Info("no dependents to heal", "master", masterName)
		return nil
	}

	h.logger.Info("waiting for master to stabilize",
		"master", masterName,
		"wait", h.stabilizeWait,
	)
	select {
	case <-time.After(h.stabilizeWait):
	case <-ctx.Done():
		return ctx.Err()
	}

	var errs []error
	for _, dep := range dependents {
		err := h.withRetry(ctx, dep.Name, func() error {
			switch classification {
			case detector.RestartCase:
				return h.healRestart(ctx, dep)
			case detector.RecreateCase:
				return h.healRecreate(ctx, dep, masterID, masterName)
			default:
				return fmt.Errorf("unknown classification: %d", classification)
			}
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("healing %s: %w", dep.Name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("healing errors: %v", errs)
	}
	return nil
}

func (h *Healer) healRestart(ctx context.Context, dep detector.Dependent) error {
	h.logger.Info("restarting dependent", "container", dep.Name)
	if err := h.client.ContainerRestart(ctx, dep.ContainerID, container.StopOptions{}); err != nil {
		return err
	}
	h.logger.Info("healed dependent", "container", dep.Name)
	return nil
}

func (h *Healer) healRecreate(ctx context.Context, dep detector.Dependent, masterID, masterName string) error {
	h.logger.Info("recreating dependent",
		"container", dep.Name,
		"new_master_id", masterID[:12],
	)

	info, err := h.client.ContainerInspect(ctx, dep.ContainerID)
	if err != nil {
		return fmt.Errorf("inspecting dependent %s: %w", dep.Name, err)
	}

	h.logger.Debug("stopping dependent", "container", dep.Name)
	if err := h.client.ContainerStop(ctx, dep.ContainerID, container.StopOptions{}); err != nil {
		return fmt.Errorf("stopping dependent %s: %w", dep.Name, err)
	}

	h.logger.Debug("removing dependent", "container", dep.Name)
	if err := h.client.ContainerRemove(ctx, dep.ContainerID, container.RemoveOptions{}); err != nil {
		return fmt.Errorf("removing dependent %s: %w", dep.Name, err)
	}

	hostConfig := info.HostConfig
	hostConfig.NetworkMode = container.NetworkMode(fmt.Sprintf("container:%s", masterID))

	cfg := info.Config
	cfg.Hostname = ""
	cfg.Domainname = ""

	containerName := dep.Name

	h.logger.Debug("creating dependent with new network mode",
		"container", containerName,
		"network_mode", string(hostConfig.NetworkMode),
	)
	createResp, err := h.client.ContainerCreate(ctx, cfg, hostConfig, nil, nil, containerName)
	if err != nil {
		return fmt.Errorf("creating dependent %s: %w", containerName, err)
	}

	if dep.WasRunning {
		h.logger.Debug("starting dependent", "container", containerName)
		if err := h.client.ContainerStart(ctx, createResp.ID, container.StartOptions{}); err != nil {
			return fmt.Errorf("starting dependent %s: %w", containerName, err)
		}
	}

	h.logger.Info("healed dependent",
		"container", containerName,
		"was_running", dep.WasRunning,
		"master", masterName,
	)
	return nil
}

func (h *Healer) withRetry(ctx context.Context, name string, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt <= h.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := h.retryBackoff * time.Duration(1<<(attempt-1))
			h.logger.Warn("retrying heal",
				"container", name,
				"attempt", attempt,
				"backoff", backoff,
				"error", lastErr,
			)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}
	}
	return fmt.Errorf("exhausted %d retries: %w", h.maxRetries, lastErr)
}
