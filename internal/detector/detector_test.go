package detector

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/st0o0/eir/internal/watcher"
)

type mockClient struct {
	containers     []container.Summary
	inspectResults map[string]container.InspectResponse
}

func (m *mockClient) ContainerList(_ context.Context, opts container.ListOptions) ([]container.Summary, error) {
	if opts.Filters.Len() > 0 {
		var filtered []container.Summary
		labelFilters := opts.Filters.Get("label")
		for _, c := range m.containers {
			info := m.inspectResults[c.ID]
			for _, lf := range labelFilters {
				parts := splitLabel(lf)
				if len(parts) == 2 && info.Config != nil && info.Config.Labels[parts[0]] == parts[1] {
					filtered = append(filtered, c)
				}
			}
		}
		return filtered, nil
	}
	return m.containers, nil
}

func splitLabel(s string) []string {
	for i, c := range s {
		if c == '=' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func (m *mockClient) ContainerInspect(_ context.Context, id string) (container.InspectResponse, error) {
	if info, ok := m.inspectResults[id]; ok {
		return info, nil
	}
	return container.InspectResponse{}, fmt.Errorf("container %s not found", id)
}

func (m *mockClient) ContainerStart(context.Context, string, container.StartOptions) error {
	return nil
}
func (m *mockClient) ContainerStop(context.Context, string, container.StopOptions) error {
	return nil
}
func (m *mockClient) ContainerRemove(context.Context, string, container.RemoveOptions) error {
	return nil
}
func (m *mockClient) ContainerCreate(context.Context, *container.Config, *container.HostConfig, *network.NetworkingConfig, *ocispec.Platform, string) (container.CreateResponse, error) {
	return container.CreateResponse{}, nil
}
func (m *mockClient) ContainerRestart(context.Context, string, container.StopOptions) error {
	return nil
}
func (m *mockClient) Events(context.Context, events.ListOptions) (<-chan events.Message, <-chan error) {
	return nil, nil
}
func (m *mockClient) Ping(context.Context) (types.Ping, error) {
	return types.Ping{}, nil
}
func (m *mockClient) Close() error { return nil }

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestClassify_DieEvent(t *testing.T) {
	d := New(&mockClient{}, "inspect", discardLogger)

	_, _, ok := d.Classify(context.Background(), watcher.Event{
		MasterName:  "vpn",
		Action:      "die",
		ContainerID: "aabbccddee112233",
		Time:        time.Now(),
	})

	if ok {
		t.Error("die events should not trigger healing")
	}
}

func TestClassify_RestartCase(t *testing.T) {
	client := &mockClient{
		containers: []container.Summary{
			{ID: "dep111111111111"},
		},
		inspectResults: map[string]container.InspectResponse{
			"dep111111111111": {
				ContainerJSONBase: &container.ContainerJSONBase{
					Name:  "/sidecar",
					State: &container.State{Running: true},
					HostConfig: &container.HostConfig{
						NetworkMode: "container:vpn",
					},
				},
				Config: &container.Config{},
			},
		},
	}

	d := New(client, "inspect", discardLogger)
	masterID := "master1111111111"

	d.lastKnownIDs["vpn"] = masterID

	classification, deps, ok := d.Classify(context.Background(), watcher.Event{
		MasterName:  "vpn",
		Action:      "start",
		ContainerID: masterID,
		Time:        time.Now(),
	})

	if !ok {
		t.Fatal("start event should trigger healing")
	}
	if classification != RestartCase {
		t.Errorf("classification = %v, want RestartCase", classification)
	}
	if len(deps) != 1 {
		t.Fatalf("deps = %d, want 1", len(deps))
	}
	if deps[0].Name != "sidecar" {
		t.Errorf("dep name = %q, want 'sidecar'", deps[0].Name)
	}
	if !deps[0].WasRunning {
		t.Error("dep should have been running")
	}
}

func TestClassify_RecreateCase(t *testing.T) {
	client := &mockClient{
		containers:     []container.Summary{},
		inspectResults: map[string]container.InspectResponse{},
	}

	d := New(client, "inspect", discardLogger)
	d.lastKnownIDs["vpn"] = "oldid11111111111"

	classification, _, ok := d.Classify(context.Background(), watcher.Event{
		MasterName:  "vpn",
		Action:      "start",
		ContainerID: "newid22222222222",
		Time:        time.Now(),
	})

	if !ok {
		t.Fatal("start event should trigger healing")
	}
	if classification != RecreateCase {
		t.Errorf("classification = %v, want RecreateCase", classification)
	}
}

func TestClassify_FirstSeen(t *testing.T) {
	client := &mockClient{
		containers:     []container.Summary{},
		inspectResults: map[string]container.InspectResponse{},
	}

	d := New(client, "inspect", discardLogger)

	classification, _, ok := d.Classify(context.Background(), watcher.Event{
		MasterName:  "vpn",
		Action:      "start",
		ContainerID: "first111111111111",
		Time:        time.Now(),
	})

	if !ok {
		t.Fatal("first start event should trigger healing")
	}
	if classification != RecreateCase {
		t.Errorf("classification = %v, want RecreateCase for first-seen", classification)
	}
}

func TestClassify_RecreateCase_DiscoversByPreviousID(t *testing.T) {
	oldMasterID := "oldmaster111111111"
	newMasterID := "newmaster222222222"

	client := &mockClient{
		containers: []container.Summary{
			{ID: "dep111111111111"},
		},
		inspectResults: map[string]container.InspectResponse{
			"dep111111111111": {
				ContainerJSONBase: &container.ContainerJSONBase{
					Name:  "/sidecar",
					State: &container.State{Running: true},
					HostConfig: &container.HostConfig{
						NetworkMode: container.NetworkMode("container:" + oldMasterID),
					},
				},
				Config: &container.Config{},
			},
		},
	}

	d := New(client, "inspect", discardLogger)
	d.lastKnownIDs["vpn"] = oldMasterID

	classification, deps, ok := d.Classify(context.Background(), watcher.Event{
		MasterName:  "vpn",
		Action:      "start",
		ContainerID: newMasterID,
		Time:        time.Now(),
	})

	if !ok {
		t.Fatal("recreate start event should trigger healing")
	}
	if classification != RecreateCase {
		t.Errorf("classification = %v, want RecreateCase", classification)
	}
	if len(deps) != 1 {
		t.Fatalf("deps = %d, want 1", len(deps))
	}
	if deps[0].Name != "sidecar" {
		t.Errorf("dep name = %q, want 'sidecar'", deps[0].Name)
	}
}

func TestDiscoverByLabel(t *testing.T) {
	client := &mockClient{
		containers: []container.Summary{
			{ID: "labeled11111111"},
		},
		inspectResults: map[string]container.InspectResponse{
			"labeled11111111": {
				ContainerJSONBase: &container.ContainerJSONBase{
					Name:  "/labeled-dep",
					State: &container.State{Running: false},
					HostConfig: &container.HostConfig{
						NetworkMode: "container:vpn",
					},
				},
				Config: &container.Config{
					Labels: map[string]string{"eir.master": "vpn"},
				},
			},
		},
	}

	d := New(client, "label", discardLogger)
	d.lastKnownIDs["vpn"] = "oldvpn111111111"

	_, deps, ok := d.Classify(context.Background(), watcher.Event{
		MasterName:  "vpn",
		Action:      "start",
		ContainerID: "newvpn222222222",
		Time:        time.Now(),
	})

	if !ok {
		t.Fatal("should trigger healing")
	}
	if len(deps) != 1 {
		t.Fatalf("deps = %d, want 1", len(deps))
	}
	if deps[0].Name != "labeled-dep" {
		t.Errorf("dep name = %q, want 'labeled-dep'", deps[0].Name)
	}
}
