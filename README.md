# eir

[![CI](https://github.com/st0o0/eir/actions/workflows/ci.yml/badge.svg)](https://github.com/st0o0/eir/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/st0o0/eir?sort=semver)](https://github.com/st0o0/eir/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Docker container network healer — automatically restores dependent containers
when their master restarts or is recreated. The Norse goddess of healing for
your container networks.

## The problem

When a "master" container (e.g. a VPN gateway like bifrost) is recreated or
restarted, every container using `network_mode: "service:<master>"` loses its
network connectivity. Docker does not reconnect dependents automatically — they
sit there with a dead network stack until someone manually restarts them.

## The solution

eir watches master containers via the Docker event stream. When a master is
recreated or restarted, eir automatically restarts the dependent containers so
they re-attach to the master's fresh network namespace. No manual intervention,
no cron hacks, no downtime.

## Quick start

```yaml
services:
  eir:
    image: ghcr.io/st0o0/eir:latest
    container_name: eir
    restart: unless-stopped
    environment:
      EIR_MASTERS: bifrost
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
```

That's it. eir will watch the `bifrost` container and restart any container
whose `network_mode` points to it whenever bifrost is recreated or restarted.

## Configuration

All configuration is via environment variables.

| Variable | Default | Purpose |
|---|---|---|
| `EIR_MASTERS` | *(required)* | Comma-separated list of master container names to watch |
| `EIR_DISCOVERY_MODE` | `inspect` | Discovery mode: `inspect` or `label` (see below) |
| `EIR_STABILIZE_WAIT` | `15s` | Wait after master start before healing dependents |
| `EIR_MAX_RETRIES` | `3` | Max retry attempts per dependent |
| `EIR_RETRY_BACKOFF` | `5s` | Initial backoff interval (exponential) |
| `EIR_LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `EIR_LOG_FORMAT` | `text` | Log format: `text` or `json` |

### Multiple masters

```yaml
environment:
  EIR_MASTERS: bifrost,gluetun,tailscale
```

## Discovery modes

eir needs to know which containers depend on each master. Two modes are
available:

### inspect (default)

eir inspects all running containers and finds those whose `NetworkMode` is
`container:<master-id>`. This is fully automatic — no labels needed.

### label

In label mode, dependents opt in with a label:

```yaml
services:
  app:
    image: yourapp:latest
    network_mode: "service:bifrost"
    labels:
      eir.master: bifrost
```

Label mode is useful when inspect is too broad or when containers use indirect
network references.

## Recovery behavior

| Master event | eir action |
|---|---|
| Master **restarted** (same container ID) | `docker restart` each dependent after `EIR_STABILIZE_WAIT` |
| Master **recreated** (new container ID) | Recreate each dependent (stop → remove → create with new NetworkMode → start) after `EIR_STABILIZE_WAIT` |
| Master **stopped** | No action — eir waits for the master to return |

eir preserves each dependent's previous state: a dependent that was stopped
before the master event is recreated but left stopped. Failed heals are retried
with exponential backoff up to `EIR_MAX_RETRIES` times.

## Image

- Registry: `ghcr.io/st0o0/eir`
- Tags: `latest`, `MAJOR.MINOR`, and the exact `MAJOR.MINOR.PATCH` per release.
- Architectures: `linux/amd64`, `linux/arm64`.

## Building from source

```bash
# build the binary
go build -o eir ./cmd/eir

# run tests
go test ./...

# vet + lint (golangci-lint v2)
go vet ./...
golangci-lint run

# build the Docker image
docker build -t eir:local .
```

Commits follow [Conventional Commits](https://www.conventionalcommits.org/);
releases and the GHCR image are cut automatically by release-please.

## License

MIT (see [`LICENSE`](LICENSE)). eir is a pure-Go static binary; the built
image bundles no third-party GPL binaries. Its Go module dependencies are
permissively licensed (MIT/Apache-2.0); see [`NOTICE`](NOTICE).
