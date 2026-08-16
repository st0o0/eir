## 1. Config

- [x] 1.1 Add `MetricsAddr` field to `Config` struct with `EIR_METRICS_ADDR` env var, default `:9550`
- [x] 1.2 Add config validation: MetricsAddr must be a valid address (`:port` or `host:port`)
- [x] 1.3 Add config tests for MetricsAddr default and custom values
- [x] 1.4 Add `MetricsAddr` to the Defaults scenario in config_test.go

## 2. Metrics Package

- [x] 2.1 Create `internal/metrics/metrics.go` with Prometheus metric definitions: `eir_events_received_total` (CounterVec, label: action), `eir_heals_total` (CounterVec, labels: master, status), `eir_heal_duration_seconds` (HistogramVec, label: master)
- [x] 2.2 Create a `Metrics` struct that holds the instruments and exposes `RecordEvent(action)`, `RecordHeal(master, error, duration)` methods
- [x] 2.3 Create `internal/metrics/server.go` with HTTP server serving `/metrics` (promhttp) and `/healthz` (health handler)
- [x] 2.4 Implement `healthzHandler` as a closure-based `http.HandlerFunc` receiving version, startTime, masters, and a ping function
- [x] 2.5 The `/healthz` handler SHALL return `{"status":"ok","version":"...","uptime":"...","masters":[...]}` on 200, or `{"status":"error","error":"..."}` on 503
- [x] 2.6 Add `Start(ctx)` method that runs the server in a goroutine with graceful shutdown on context cancellation
- [x] 2.7 Add unit tests for `healthzHandler` (healthy response, docker unreachable)
- [x] 2.8 Add unit tests for `RecordEvent` and `RecordHeal` (verify metric values via testutil)

## 3. Integrate into Main

- [x] 3.1 Create metrics instance in `main.go` and start the metrics server before the event loop
- [x] 3.2 Call `metrics.RecordEvent(event.Action)` on each received event
- [x] 3.3 Call `metrics.RecordHeal(master, err, duration)` after each heal operation (wrap `h.Heal` with timing)
- [x] 3.4 Log metrics server address at startup
- [x] 3.5 Remove `touchHealthFile()` calls from the event loop
- [x] 3.6 Remove the 10-second heartbeat ticker from the event loop
- [x] 3.7 Delete `cmd/eir/health.go` (heartbeat file writer)

## 4. Rewrite Healthcheck Subcommand

- [x] 4.1 Rewrite `cmd/eir/healthcheck.go`: load config, derive localhost address from MetricsAddr, HTTP GET `/healthz`
- [x] 4.2 Implement `healthAddr()` helper: `:9550` → `localhost:9550`, `0.0.0.0:9550` → `localhost:9550`, already-qualified stays as-is
- [x] 4.3 Implement `doHealthcheck(addr)` for testability: HTTP GET with 5s timeout, exit 0 on 200, exit 1 otherwise
- [x] 4.4 Add unit tests for `healthAddr()` (port-only, wildcard, qualified)
- [x] 4.5 Add unit tests for `doHealthcheck()` (200 → 0, 503 → 1, connection refused → 1)
- [x] 4.6 Remove old healthcheck tests that test file-based logic

## 5. Dockerfile & Compose

- [x] 5.1 Update Dockerfile: `EXPOSE 9550`, adjust `--start-period` to match ran (`15s`)
- [x] 5.2 Update docker-compose.yml: add port mapping `9550:9550`
- [x] 5.3 Build and verify: healthcheck passes in running scratch container
