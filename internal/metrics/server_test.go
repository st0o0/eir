package metrics

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthzHandler_Healthy(t *testing.T) {
	startTime := time.Now().Add(-2 * time.Hour)
	masters := []string{"bifrost", "gluetun"}
	ping := func(ctx context.Context) error { return nil }

	handler := healthzHandler("1.0.0", startTime, masters, ping)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var resp healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("status = %q, want 'ok'", resp.Status)
	}
	if resp.Version != "1.0.0" {
		t.Errorf("version = %q, want '1.0.0'", resp.Version)
	}
	if len(resp.Masters) != 2 {
		t.Errorf("masters = %v, want [bifrost gluetun]", resp.Masters)
	}
	if resp.Uptime == "" {
		t.Error("uptime is empty")
	}
}

func TestHealthzHandler_DockerUnreachable(t *testing.T) {
	ping := func(ctx context.Context) error { return errors.New("connection refused") }

	handler := healthzHandler("1.0.0", time.Now(), []string{"bifrost"}, ping)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}

	var resp healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("status = %q, want 'error'", resp.Status)
	}
	if resp.Error == "" {
		t.Error("error message is empty")
	}
}
