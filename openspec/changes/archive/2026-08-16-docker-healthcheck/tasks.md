## 1. Heartbeat Writer

- [x] 1.1 Create `cmd/eir/health.go` with shared constant `healthFilePath = "/eir.health"` and `touchHealthFile()` function that writes the current Unix timestamp to the file
- [x] 1.2 Add a 10-second ticker to the event loop in `cmd/eir/main.go` that calls `touchHealthFile()`
- [x] 1.3 Call `touchHealthFile()` after each processed event in the main loop
- [x] 1.4 Ensure `touchHealthFile()` is NOT called before the watcher channels are established (startup tracking)

## 2. Healthcheck Subcommand

- [x] 2.1 Rewrite `cmd/eir/healthcheck.go` to check heartbeat file existence (exit 1 + "not ready" if missing)
- [x] 2.2 Add heartbeat file staleness check (exit 1 + "stale heartbeat" if age > 60s)
- [x] 2.3 Keep Docker daemon ping as final check (exit 1 + "docker unreachable" on failure)
- [x] 2.4 Return exit 0 only when all three checks pass

## 3. Tests

- [x] 3.1 Unit test `touchHealthFile()` — file is created, content is a timestamp, mtime updates on repeated calls
- [x] 3.2 Unit test healthcheck: file missing → exit 1
- [x] 3.3 Unit test healthcheck: file stale (age > 60s) → exit 1
- [x] 3.4 Unit test healthcheck: file fresh + Docker ping fails → exit 1
- [x] 3.5 Unit test healthcheck: file fresh + Docker ping ok → exit 0

## 4. Dockerfile & Integration

- [x] 4.1 Update Dockerfile `HEALTHCHECK` to `--interval=30s --timeout=10s --start-period=45s --retries=3` (aligned with bifrost)
- [x] 4.2 Verify scratch image supports writing to `/eir.health` (build and run locally)
