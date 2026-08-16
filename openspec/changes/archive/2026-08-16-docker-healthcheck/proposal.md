## Why

The current health check only pings the Docker daemon — it cannot detect when eir's own event loop hangs, the watcher channel closes, or startup hasn't completed yet. Docker will report eir as "healthy" even when it has silently stopped processing events. A heartbeat-file approach lets the health check verify that the main loop is alive and that startup succeeded.

## What Changes

- Introduce a heartbeat file (`/eir.health`) written by the main event loop on a 10-second ticker and on every processed event
- Rewrite the `healthcheck` subcommand to check heartbeat file existence (startup), staleness (liveness), and Docker daemon reachability — in that order
- The health file is not written until the watcher has successfully connected, giving explicit startup tracking
- Align Dockerfile `HEALTHCHECK` parameters with bifrost (`--start-period=45s`)

## Capabilities

### New Capabilities

_(none — this extends an existing capability)_

### Modified Capabilities

- `eir`: The health check requirement changes from "ping Docker daemon" to a three-phase check: startup readiness (file exists), loop liveness (file age < 60s), and Docker daemon reachability

## Impact

- **Code**: `cmd/eir/main.go` (add ticker + heartbeat writes), `cmd/eir/healthcheck.go` (rewrite checks), new `cmd/eir/health.go` (shared constant + write function)
- **Dockerfile**: Possibly adjust `--start-period`
- **Runtime**: Health file written to container filesystem at `/eir.health` — no volume mount needed, scratch image supports writable overlay
- **No breaking changes**: The healthcheck subcommand interface (`/eir healthcheck` → exit 0/1) stays identical
