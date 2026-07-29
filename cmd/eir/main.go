package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/client"

	"github.com/st0o0/eir/internal/config"
	"github.com/st0o0/eir/internal/detector"
	"github.com/st0o0/eir/internal/healer"
	"github.com/st0o0/eir/internal/watcher"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthcheck())
	}

	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println(version)
		os.Exit(0)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}

	logger := newLogger(cfg)

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logger.Error("failed to create docker client", "error", err)
		os.Exit(1)
	}
	defer dockerClient.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("eir starting",
		"version", version,
		"masters", cfg.Masters,
		"discovery_mode", cfg.DiscoveryMode,
	)

	w := watcher.New(dockerClient, cfg.Masters, logger)
	det := detector.New(dockerClient, cfg.DiscoveryMode, logger)
	h := healer.New(dockerClient, cfg.StabilizeWait, cfg.MaxRetries, cfg.RetryBackoff, logger)

	eventCh, errCh := w.Watch(ctx)

	for {
		select {
		case <-ctx.Done():
			logger.Info("shutting down")
			return
		case err, ok := <-errCh:
			if !ok {
				logger.Error("event stream closed unexpectedly")
				return
			}
			logger.Error("watcher error", "error", err)
			return
		case event, ok := <-eventCh:
			if !ok {
				logger.Info("event channel closed")
				return
			}

			classification, dependents, shouldHeal := det.Classify(ctx, event)
			if !shouldHeal {
				continue
			}

			logger.Info("healing dependents",
				"master", event.MasterName,
				"classification", classification.String(),
				"dependents", len(dependents),
			)

			if err := h.Heal(ctx, classification, event.ContainerID, event.MasterName, dependents); err != nil {
				logger.Error("healing failed",
					"master", event.MasterName,
					"error", err,
				)
			}
		}
	}
}

func newLogger(cfg *config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	var handler slog.Handler
	if cfg.LogFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}
