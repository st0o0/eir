## Context

eir runs as a Docker sidecar in a `scratch`-based container. It has no HTTP server — health is checked via a CLI subcommand (`/eir healthcheck`) invoked by Docker's `HEALTHCHECK` directive. The current implementation only pings the Docker daemon, which cannot detect internal failures like a hung event loop or incomplete startup.

## Goals / Non-Goals

**Goals:**

- Detect when the main event loop stops processing (liveness)
- Detect when startup hasn't completed yet (readiness)
- Keep Docker daemon reachability check (infrastructure)
- Zero additional dependencies — no HTTP server, no Unix socket

**Non-Goals:**

- Exposing health via HTTP endpoint (no web server in eir)
- Making the health file path configurable (internal detail)
- Distinguishing between different types of Docker daemon failures

## Decisions

### Heartbeat file over Unix socket or HTTP

The main loop writes a timestamp to `/eir.health` every 10 seconds via a ticker, plus on every processed event. The healthcheck subcommand reads the file's modification time.

**Why:** A file is the simplest IPC mechanism that works in a scratch image. No listeners, no connection lifecycle, no extra goroutines for serving. The healthcheck process just calls `os.Stat()`.

**Alternatives considered:**
- Unix socket: more complex lifecycle management, unclear behavior in scratch
- HTTP endpoint: requires a listener goroutine, port management, and adds attack surface

### Check order: file existence → file age → Docker ping

The healthcheck runs the cheapest checks first. If the file doesn't exist, we're still starting up — no point pinging Docker. If the file is stale, the loop is dead — Docker connectivity doesn't matter.

### Stale threshold: 60 seconds

The ticker writes every 10s. The Docker HEALTHCHECK runs every 30s with 3 retries. A 60s threshold means:
- Normal jitter won't cause false positives (6 missed ticks before stale)
- A truly hung loop is detected within ~90s (60s stale + one health check cycle)

### Health file path: `/eir.health`

Written directly to the container root filesystem. The scratch image's overlay is writable. No `/tmp` directory needed, no volume mount needed.

### Startup tracking via file absence

The health file is not created until the watcher has successfully established the Docker event stream. During startup, the file simply doesn't exist, and the healthcheck reports "not ready". Docker's `--start-period=45s` (aligned with bifrost) prevents this from counting as unhealthy.

### Dockerfile HEALTHCHECK aligned with bifrost

Use `--interval=30s --timeout=10s --start-period=45s --retries=3` to match bifrost's health check configuration. The longer start-period (45s vs current 15s) gives more headroom for slow Docker daemon connections or large event filter setups.

## Risks / Trade-offs

- **[File I/O on every tick]** → Negligible: one `os.WriteFile` of a few bytes every 10 seconds
- **[Clock skew in container]** → Mitigated: both writer and reader use the same system clock via `os.Stat` mtime
- **[scratch image writable overlay]** → Docker guarantees a writable overlay by default; read-only rootfs would need an explicit volume, but that's an edge case eir doesn't target
