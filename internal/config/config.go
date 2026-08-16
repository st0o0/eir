package config

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all eir runtime configuration parsed from environment variables.
type Config struct {
	Masters        []string
	DiscoveryMode  string
	StabilizeWait  time.Duration
	MaxRetries     int
	RetryBackoff   time.Duration
	LogLevel       slog.Level
	LogFormat      string
	MetricsAddr    string
}

// Load parses EIR_* environment variables into a Config.
func Load() (*Config, error) {
	masters := os.Getenv("EIR_MASTERS")
	if masters == "" {
		return nil, fmt.Errorf("EIR_MASTERS is required")
	}

	var masterList []string
	for _, m := range strings.Split(masters, ",") {
		m = strings.TrimSpace(m)
		if m != "" {
			masterList = append(masterList, m)
		}
	}
	if len(masterList) == 0 {
		return nil, fmt.Errorf("EIR_MASTERS must contain at least one container name")
	}

	discoveryMode := envOrDefault("EIR_DISCOVERY_MODE", "inspect")
	if discoveryMode != "inspect" && discoveryMode != "label" {
		return nil, fmt.Errorf("EIR_DISCOVERY_MODE must be 'inspect' or 'label', got %q", discoveryMode)
	}

	stabilizeWait, err := parseDuration("EIR_STABILIZE_WAIT", "15s")
	if err != nil {
		return nil, err
	}

	maxRetries, err := parseInt("EIR_MAX_RETRIES", "3")
	if err != nil {
		return nil, err
	}

	retryBackoff, err := parseDuration("EIR_RETRY_BACKOFF", "5s")
	if err != nil {
		return nil, err
	}

	logLevel, err := parseLogLevel(envOrDefault("EIR_LOG_LEVEL", "info"))
	if err != nil {
		return nil, err
	}

	logFormat := envOrDefault("EIR_LOG_FORMAT", "text")
	if logFormat != "text" && logFormat != "json" {
		return nil, fmt.Errorf("EIR_LOG_FORMAT must be 'text' or 'json', got %q", logFormat)
	}

	metricsAddr := envOrDefault("EIR_METRICS_ADDR", ":9550")
	if _, _, err := net.SplitHostPort(metricsAddr); err != nil {
		return nil, fmt.Errorf("invalid EIR_METRICS_ADDR %q: %w", metricsAddr, err)
	}

	return &Config{
		Masters:       masterList,
		DiscoveryMode: discoveryMode,
		StabilizeWait: stabilizeWait,
		MaxRetries:    maxRetries,
		RetryBackoff:  retryBackoff,
		LogLevel:      logLevel,
		LogFormat:     logFormat,
		MetricsAddr:   metricsAddr,
	}, nil
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func parseDuration(envKey, defaultVal string) (time.Duration, error) {
	raw := envOrDefault(envKey, defaultVal)
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value %q: %w", envKey, raw, err)
	}
	return d, nil
}

func parseInt(envKey, defaultVal string) (int, error) {
	raw := envOrDefault(envKey, defaultVal)
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value %q: %w", envKey, raw, err)
	}
	return n, nil
}

func parseLogLevel(s string) (slog.Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid EIR_LOG_LEVEL %q: must be debug, info, warn, or error", s)
	}
}
