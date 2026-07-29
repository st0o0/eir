package watcher

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/docker/docker/api/types/events"

	dockerclient "github.com/st0o0/eir/internal/docker"
)

// Event represents a relevant Docker container lifecycle event.
type Event struct {
	MasterName  string
	Action      string
	ContainerID string
	Time        time.Time
}

// Watcher subscribes to Docker events for configured master containers.
type Watcher struct {
	client  dockerclient.Client
	masters []string
	logger  *slog.Logger
}

// New creates a Watcher for the given master container names.
func New(client dockerclient.Client, masters []string, logger *slog.Logger) *Watcher {
	return &Watcher{
		client:  client,
		masters: masters,
		logger:  logger,
	}
}

// Watch starts listening for Docker events and returns channels for events and errors.
func (w *Watcher) Watch(ctx context.Context) (<-chan Event, <-chan error) {
	out := make(chan Event)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		filter := dockerclient.EventFilter(w.masters)
		msgCh, dockerErrCh := w.client.Events(ctx, events.ListOptions{Filters: filter})

		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-dockerErrCh:
				if !ok {
					return
				}
				errCh <- fmt.Errorf("docker event stream: %w", err)
				return
			case msg, ok := <-msgCh:
				if !ok {
					return
				}
				ev := Event{
					MasterName:  msg.Actor.Attributes["name"],
					Action:      string(msg.Action),
					ContainerID: msg.Actor.ID,
					Time:        time.Unix(msg.Time, msg.TimeNano),
				}
				w.logger.Info("received event",
					"master", ev.MasterName,
					"action", ev.Action,
					"container_id", ev.ContainerID[:12],
				)
				select {
				case out <- ev:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, errCh
}
