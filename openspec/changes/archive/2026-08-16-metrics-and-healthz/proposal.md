## Why

eir has zero observability beyond structured logs. There are no metrics to monitor healing activity, event throughput, or failure rates. The current file-based healthcheck works but is an indirect signal — it cannot report structured state like version, uptime, or watched masters. Adding a metrics server with a `/healthz` endpoint (following the ran pattern) gives both Prometheus scraping and a richer health probe in one component.

## What Changes

- Add an HTTP metrics server on a configurable address (default `:9550`) serving `/metrics` (Prometheus) and `/healthz` (JSON health response)
- Instrument the event loop with Prometheus counters and histograms: events received, heals performed (success/fail), heal duration
- Replace the file-based healthcheck with an HTTP-based approach: the `healthcheck` subcommand GETs `/healthz` from the metrics server instead of checking a heartbeat file
- The `/healthz` handler checks Docker daemon reachability inline and returns a JSON response with status, version, uptime, and master list
- Remove the heartbeat file writer (`touchHealthFile`, health file, ticker) — liveness is now implicit: HTTP responds = process lives
- Add `EIR_METRICS_ADDR` config variable (default `:9550`)

## Capabilities

### New Capabilities

- `metrics`: Prometheus metrics endpoint and instrumented counters/histograms for event processing and healing operations

### Modified Capabilities

- `eir`: Health check requirement changes from file-based heartbeat to HTTP-based `/healthz` endpoint; new config variable `EIR_METRICS_ADDR`

## Impact

- **Code**: New `internal/metrics` package, rewrite `cmd/eir/healthcheck.go` and `cmd/eir/health.go`, remove heartbeat file logic from `cmd/eir/main.go`, extend `internal/config`
- **Dependencies**: Add `github.com/prometheus/client_golang` as direct dependency
- **Dockerfile**: Adjust `HEALTHCHECK` start-period, expose port 9550
- **Docker Compose**: Add port mapping for metrics
- **No breaking changes**: The `healthcheck` subcommand interface stays the same (`/eir healthcheck` → exit 0/1)
