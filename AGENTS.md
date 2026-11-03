# eir

Docker container network healer. Automatically restores dependent containers when their master (e.g. a VPN gateway) restarts or is recreated.

## Architecture

```
cmd/eir/            Entry point, healthcheck, signal handling
internal/
  config/           EIR_ env vars, validated at startup
  docker/           Docker SDK client interface + event filter
  watcher/          Docker event stream subscription for master containers
  detector/         Event classification (restart vs recreate) + dependent discovery
  healer/           Recovery actions: restart or stop/remove/create/start with retry
  metrics/          Prometheus /metrics endpoint (eir_ prefix), /healthz
```

## Build

```bash
go build -o eir ./cmd/eir
go test ./...
go vet ./...
golangci-lint run
```

## Conventions

- Go 1.26+, pure Go, no CGO
- Conventional commits: feat/fix/perf/docs/chore/refactor/test/ci/build/deps
- commitlint enforced, release-please for versioning
- golangci-lint v2, hadolint for Dockerfile
- All env vars prefixed with `EIR_`
- Structured logging via log/slog (JSON or text)
- Metric names prefixed with `eir_`
- Custom Prometheus registry (never DefaultRegisterer)
- Errors wrapped with `fmt.Errorf` and `%w`

## Deployment

Single container with read-only Docker socket mount. Deployed via compose alongside the master container it watches. Published as `ghcr.io/st0o0/eir`.

## Security-sensitive areas

- `internal/docker/` -- Docker socket access (read-only)
- `internal/healer/` -- container lifecycle operations (stop, remove, create, start)
- `internal/detector/` -- inspects all containers to discover dependents
