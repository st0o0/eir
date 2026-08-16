## Context

eir is a pure event-loop Docker sidecar running in a scratch container. It currently has no HTTP server and no metrics. The healthcheck uses a file-based heartbeat. The ran project has a proven pattern: a metrics server on `:9550` serving both `/metrics` (Prometheus) and `/healthz` (JSON health), with the healthcheck subcommand querying `/healthz` via HTTP.

## Goals / Non-Goals

**Goals:**

- Prometheus metrics for event processing and healing operations
- HTTP-based health endpoint with structured JSON response
- Replace file-based healthcheck with HTTP-based (ran pattern)
- Consistent observability approach across ran and eir

**Non-Goals:**

- OpenTelemetry tracing (overkill for a sidecar with one event loop)
- Custom metric dashboards or Grafana provisioning
- Metrics authentication or TLS

## Decisions

### Package structure: `internal/metrics`

A single `internal/metrics` package owns the HTTP server, Prometheus registry, metric definitions, and the `/healthz` handler. This keeps all observability concerns in one place.

**Why:** The metrics server is a cross-cutting concern that touches events, healing, and health. A dedicated package prevents metric definitions from leaking into domain packages.

### Prometheus client library over OpenTelemetry metrics

Use `github.com/prometheus/client_golang` directly rather than OTel metrics SDK with a Prometheus exporter.

**Why:** eir only needs Prometheus exposition. The OTel metrics SDK adds a provider, reader, and exporter layer that's unnecessary here. The Prometheus client is battle-tested, zero-config for `/metrics`, and half the dependency weight. The OTel modules in go.mod are indirect (from Docker SDK) — no reason to promote them.

### `/healthz` handler gets dependencies via closure

The handler receives version, start time, master list, and a Docker ping function via closure at construction time — no global state, same pattern as ran.

```go
func healthzHandler(version string, startTime time.Time, masters []string, ping func(ctx) error) http.HandlerFunc
```

**Why:** Testable without a running server. The ping function can be swapped for a mock in tests.

### Metric instruments

| Metric | Type | Labels | Package |
|--------|------|--------|---------|
| `eir_events_received_total` | CounterVec | `action` | metrics |
| `eir_heals_total` | CounterVec | `master`, `status` | metrics |
| `eir_heal_duration_seconds` | HistogramVec | `master` | metrics |

Heal duration buckets: `{0.5, 1, 2, 5, 10, 30, 60}` — healing involves container restart/recreate, typically 1-30s.

**Why:** Minimal set that answers "is eir doing its job?" — event throughput, heal success rate, heal latency. More can be added later.

### Healthcheck subcommand: HTTP GET replaces file check

The `healthcheck` subcommand loads config to get `EIR_METRICS_ADDR`, derives a localhost address (`:9550` → `localhost:9550`), and does `HTTP GET /healthz`. Exit 0 on 200, exit 1 otherwise.

**Why:** The HTTP server being alive IS the liveness signal. No separate heartbeat file needed. If the event loop is stuck, the HTTP server goroutine still responds — but that's acceptable because a stuck event loop would show up in stale metrics (no events processed, no heals). The Docker ping in the `/healthz` handler catches the case where the daemon is gone.

### Remove file-based healthcheck entirely

Delete `touchHealthFile()`, the health file constant, the ticker in the event loop, and all file-check logic from `runHealthcheck()`.

**Why:** The HTTP-based approach is strictly better — it provides structured health info, doesn't need file I/O, and works the same in scratch images.

### Metrics server lifecycle

Start the HTTP server in a goroutine before entering the event loop. Use `http.Server` with `Shutdown()` on context cancellation for graceful cleanup.

**Why:** The server must be ready before Docker's HEALTHCHECK start-period expires. Starting it early (before the watcher) means the healthcheck subcommand gets a connection even during startup — and the `/healthz` handler can still report Docker unreachable if the daemon isn't ready yet.

## Risks / Trade-offs

- **[New dependency: prometheus/client_golang]** → Well-maintained, widely used, small footprint. Acceptable for the value it provides.
- **[HTTP server in scratch image]** → Go's `net/http` is statically linked in, no additional binaries needed. Verified with the existing scratch build.
- **[Stuck event loop not directly detected]** → The HTTP server runs in a separate goroutine and would still respond. Mitigation: stale metrics (flat counters) serve as an indirect signal. The file-based heartbeat was the only direct liveness check, but in practice a stuck goroutine is rare in Go, and the Docker ping covers the most common failure mode.
- **[Port conflict with other services]** → Mitigated by making the address configurable via `EIR_METRICS_ADDR`.
