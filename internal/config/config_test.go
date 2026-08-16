package config

import (
	"log/slog"
	"os"
	"testing"
	"time"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"EIR_MASTERS", "EIR_DISCOVERY_MODE", "EIR_STABILIZE_WAIT",
		"EIR_MAX_RETRIES", "EIR_RETRY_BACKOFF", "EIR_LOG_LEVEL", "EIR_LOG_FORMAT",
		"EIR_METRICS_ADDR",
	} {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}
}

func TestLoad_Defaults(t *testing.T) {
	clearEnv(t)
	t.Setenv("EIR_MASTERS", "nginx,postgres")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Masters) != 2 || cfg.Masters[0] != "nginx" || cfg.Masters[1] != "postgres" {
		t.Errorf("masters = %v, want [nginx postgres]", cfg.Masters)
	}
	if cfg.DiscoveryMode != "inspect" {
		t.Errorf("discovery mode = %q, want 'inspect'", cfg.DiscoveryMode)
	}
	if cfg.StabilizeWait != 15*time.Second {
		t.Errorf("stabilize wait = %v, want 15s", cfg.StabilizeWait)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("max retries = %d, want 3", cfg.MaxRetries)
	}
	if cfg.RetryBackoff != 5*time.Second {
		t.Errorf("retry backoff = %v, want 5s", cfg.RetryBackoff)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("log level = %v, want info", cfg.LogLevel)
	}
	if cfg.LogFormat != "text" {
		t.Errorf("log format = %q, want 'text'", cfg.LogFormat)
	}
	if cfg.MetricsAddr != ":9550" {
		t.Errorf("metrics addr = %q, want ':9550'", cfg.MetricsAddr)
	}
}

func TestLoad_MissingMasters(t *testing.T) {
	clearEnv(t)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing EIR_MASTERS")
	}
}

func TestLoad_EmptyMasters(t *testing.T) {
	clearEnv(t)
	t.Setenv("EIR_MASTERS", " , , ")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for empty EIR_MASTERS")
	}
}

func TestLoad_InvalidDiscoveryMode(t *testing.T) {
	clearEnv(t)
	t.Setenv("EIR_MASTERS", "nginx")
	t.Setenv("EIR_DISCOVERY_MODE", "magic")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid discovery mode")
	}
}

func TestLoad_LabelMode(t *testing.T) {
	clearEnv(t)
	t.Setenv("EIR_MASTERS", "vpn")
	t.Setenv("EIR_DISCOVERY_MODE", "label")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DiscoveryMode != "label" {
		t.Errorf("discovery mode = %q, want 'label'", cfg.DiscoveryMode)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	clearEnv(t)
	t.Setenv("EIR_MASTERS", "wireguard")
	t.Setenv("EIR_STABILIZE_WAIT", "30s")
	t.Setenv("EIR_MAX_RETRIES", "5")
	t.Setenv("EIR_RETRY_BACKOFF", "10s")
	t.Setenv("EIR_LOG_LEVEL", "debug")
	t.Setenv("EIR_LOG_FORMAT", "json")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.StabilizeWait != 30*time.Second {
		t.Errorf("stabilize wait = %v, want 30s", cfg.StabilizeWait)
	}
	if cfg.MaxRetries != 5 {
		t.Errorf("max retries = %d, want 5", cfg.MaxRetries)
	}
	if cfg.RetryBackoff != 10*time.Second {
		t.Errorf("retry backoff = %v, want 10s", cfg.RetryBackoff)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("log level = %v, want debug", cfg.LogLevel)
	}
	if cfg.LogFormat != "json" {
		t.Errorf("log format = %q, want 'json'", cfg.LogFormat)
	}
}

func TestLoad_InvalidDuration(t *testing.T) {
	clearEnv(t)
	t.Setenv("EIR_MASTERS", "nginx")
	t.Setenv("EIR_STABILIZE_WAIT", "notaduration")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestLoad_InvalidRetries(t *testing.T) {
	clearEnv(t)
	t.Setenv("EIR_MASTERS", "nginx")
	t.Setenv("EIR_MAX_RETRIES", "abc")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid retries")
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	clearEnv(t)
	t.Setenv("EIR_MASTERS", "nginx")
	t.Setenv("EIR_LOG_LEVEL", "verbose")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid log level")
	}
}

func TestLoad_InvalidLogFormat(t *testing.T) {
	clearEnv(t)
	t.Setenv("EIR_MASTERS", "nginx")
	t.Setenv("EIR_LOG_FORMAT", "yaml")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid log format")
	}
}

func TestLoad_CustomMetricsAddr(t *testing.T) {
	clearEnv(t)
	t.Setenv("EIR_MASTERS", "nginx")
	t.Setenv("EIR_METRICS_ADDR", "0.0.0.0:8080")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MetricsAddr != "0.0.0.0:8080" {
		t.Errorf("metrics addr = %q, want '0.0.0.0:8080'", cfg.MetricsAddr)
	}
}

func TestLoad_InvalidMetricsAddr(t *testing.T) {
	clearEnv(t)
	t.Setenv("EIR_MASTERS", "nginx")
	t.Setenv("EIR_METRICS_ADDR", "not-an-addr")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid metrics addr")
	}
}
