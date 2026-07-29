package healer

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

	"github.com/st0o0/eir/internal/detector"
)

type call struct {
	method string
	id     string
}

type mockClient struct {
	calls          []call
	inspectResult  container.InspectResponse
	createResponse container.CreateResponse
	failOn         string
}

func (m *mockClient) ContainerList(_ context.Context, _ container.ListOptions) ([]container.Summary, error) {
	m.calls = append(m.calls, call{"list", ""})
	return nil, nil
}

func (m *mockClient) ContainerInspect(_ context.Context, id string) (container.InspectResponse, error) {
	m.calls = append(m.calls, call{"inspect", id})
	if m.failOn == "inspect" {
		return container.InspectResponse{}, fmt.Errorf("inspect failed")
	}
	return m.inspectResult, nil
}

func (m *mockClient) ContainerStart(_ context.Context, id string, _ container.StartOptions) error {
	m.calls = append(m.calls, call{"start", id})
	if m.failOn == "start" {
		return fmt.Errorf("start failed")
	}
	return nil
}

func (m *mockClient) ContainerStop(_ context.Context, id string, _ container.StopOptions) error {
	m.calls = append(m.calls, call{"stop", id})
	if m.failOn == "stop" {
		return fmt.Errorf("stop failed")
	}
	return nil
}

func (m *mockClient) ContainerRemove(_ context.Context, id string, _ container.RemoveOptions) error {
	m.calls = append(m.calls, call{"remove", id})
	if m.failOn == "remove" {
		return fmt.Errorf("remove failed")
	}
	return nil
}

func (m *mockClient) ContainerCreate(_ context.Context, _ *container.Config, _ *container.HostConfig, _ *network.NetworkingConfig, _ *ocispec.Platform, name string) (container.CreateResponse, error) {
	m.calls = append(m.calls, call{"create", name})
	if m.failOn == "create" {
		return container.CreateResponse{}, fmt.Errorf("create failed")
	}
	return m.createResponse, nil
}

func (m *mockClient) ContainerRestart(_ context.Context, id string, _ container.StopOptions) error {
	m.calls = append(m.calls, call{"restart", id})
	if m.failOn == "restart" {
		return fmt.Errorf("restart failed")
	}
	return nil
}

func (m *mockClient) Events(_ context.Context, _ events.ListOptions) (<-chan events.Message, <-chan error) {
	return nil, nil
}

func (m *mockClient) Ping(_ context.Context) (types.Ping, error) {
	return types.Ping{}, nil
}

func (m *mockClient) Close() error { return nil }

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestHeal_RestartCase(t *testing.T) {
	client := &mockClient{}
	h := New(client, 0, 0, time.Millisecond, discardLogger)

	err := h.Heal(context.Background(), detector.RestartCase, "master123456789", "vpn", []detector.Dependent{
		{ContainerID: "dep111111111111", Name: "sidecar", WasRunning: true},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(client.calls))
	}
	if client.calls[0].method != "restart" {
		t.Errorf("method = %q, want 'restart'", client.calls[0].method)
	}
}

func TestHeal_RecreateCase_WasRunning(t *testing.T) {
	client := &mockClient{
		inspectResult: container.InspectResponse{
			ContainerJSONBase: &container.ContainerJSONBase{
				HostConfig: &container.HostConfig{
					NetworkMode: "container:oldmaster",
				},
			},
			Config: &container.Config{Image: "myapp:latest"},
		},
		createResponse: container.CreateResponse{ID: "newdep999999"},
	}
	h := New(client, 0, 0, time.Millisecond, discardLogger)

	err := h.Heal(context.Background(), detector.RecreateCase, "newmaster12345678", "vpn", []detector.Dependent{
		{ContainerID: "dep111111111111", Name: "sidecar", WasRunning: true},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"inspect", "stop", "remove", "create", "start"}
	if len(client.calls) != len(expected) {
		t.Fatalf("calls = %v, want %v", methodNames(client.calls), expected)
	}
	for i, e := range expected {
		if client.calls[i].method != e {
			t.Errorf("call[%d] = %q, want %q", i, client.calls[i].method, e)
		}
	}
}

func TestHeal_RecreateCase_WasNotRunning(t *testing.T) {
	client := &mockClient{
		inspectResult: container.InspectResponse{
			ContainerJSONBase: &container.ContainerJSONBase{
				HostConfig: &container.HostConfig{
					NetworkMode: "container:oldmaster",
				},
			},
			Config: &container.Config{Image: "myapp:latest"},
		},
		createResponse: container.CreateResponse{ID: "newdep999999"},
	}
	h := New(client, 0, 0, time.Millisecond, discardLogger)

	err := h.Heal(context.Background(), detector.RecreateCase, "newmaster12345678", "vpn", []detector.Dependent{
		{ContainerID: "dep111111111111", Name: "sidecar", WasRunning: false},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should NOT start the container since it wasn't running
	expected := []string{"inspect", "stop", "remove", "create"}
	if len(client.calls) != len(expected) {
		t.Fatalf("calls = %v, want %v", methodNames(client.calls), expected)
	}
}

func TestHeal_NoDependents(t *testing.T) {
	client := &mockClient{}
	h := New(client, 0, 0, time.Millisecond, discardLogger)

	err := h.Heal(context.Background(), detector.RestartCase, "master123456789", "vpn", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.calls) != 0 {
		t.Errorf("expected no calls, got %v", client.calls)
	}
}

func TestHeal_RetryOnFailure(t *testing.T) {
	failClient := &mockClient{failOn: "restart"}
	hFail := New(failClient, 0, 1, time.Millisecond, discardLogger)

	err := hFail.Heal(context.Background(), detector.RestartCase, "master123456789", "vpn", []detector.Dependent{
		{ContainerID: "dep111111111111", Name: "sidecar", WasRunning: true},
	})

	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	// Should have tried 2 times (initial + 1 retry)
	restartCount := 0
	for _, c := range failClient.calls {
		if c.method == "restart" {
			restartCount++
		}
	}
	if restartCount != 2 {
		t.Errorf("restart attempts = %d, want 2", restartCount)
	}
}

func TestHeal_ContextCancellation(t *testing.T) {
	client := &mockClient{}
	h := New(client, 5*time.Second, 0, time.Millisecond, discardLogger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := h.Heal(ctx, detector.RestartCase, "master123456789", "vpn", []detector.Dependent{
		{ContainerID: "dep111111111111", Name: "sidecar", WasRunning: true},
	})

	if err != context.Canceled {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}

func methodNames(calls []call) []string {
	names := make([]string, len(calls))
	for i, c := range calls {
		names[i] = c.method
	}
	return names
}
