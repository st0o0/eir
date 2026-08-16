# eir

## Purpose

eir is a Docker sidecar that automatically restores dependent containers when their master container restarts or is recreated. It solves the problem that Docker does not reconnect containers using `network_mode: "container:<master>"` when the master's network namespace changes.

## Non-goals

- eir does NOT manage container orchestration beyond network healing
- eir does NOT modify container configurations permanently
- eir does NOT handle multi-host / Swarm / Kubernetes scenarios
- eir does NOT restart dependents when a master is intentionally stopped

## Requirements

### Requirement: Master watching

The system SHALL watch one or more named master containers for lifecycle events via the Docker event stream.

#### Scenario: Single master

- **GIVEN** `EIR_MASTERS=bifrost`
- **WHEN** eir starts
- **THEN** it subscribes to Docker events filtered to the `bifrost` container

#### Scenario: Multiple masters

- **GIVEN** `EIR_MASTERS=bifrost,gluetun,tailscale`
- **WHEN** eir starts
- **THEN** it watches all three masters independently

### Requirement: Configuration via environment variables

The system SHALL be configured entirely through `EIR_*` environment variables.

#### Scenario: Required masters variable

- **WHEN** `EIR_MASTERS` is not set or empty
- **THEN** eir exits with an error

#### Scenario: Defaults

- **WHEN** only `EIR_MASTERS` is set
- **THEN** the following defaults apply:
  - `EIR_DISCOVERY_MODE` = `inspect`
  - `EIR_STABILIZE_WAIT` = `15s`
  - `EIR_MAX_RETRIES` = `3`
  - `EIR_RETRY_BACKOFF` = `5s`
  - `EIR_LOG_LEVEL` = `info`
  - `EIR_LOG_FORMAT` = `text`
  - `EIR_METRICS_ADDR` = `:9550`

### Requirement: Health check

The system SHALL provide a health check via the `healthcheck` subcommand that queries the `/healthz` HTTP endpoint on the metrics server.

The `healthcheck` subcommand SHALL:

1. Load config to determine the metrics address
2. Derive `localhost:<port>` from the configured bind address
3. HTTP GET `http://localhost:<port>/healthz` with a 5-second timeout
4. Exit 0 on HTTP 200, exit 1 on any other response or error

#### Scenario: Healthy

- **WHEN** the healthcheck subcommand runs and `/healthz` returns HTTP 200
- **THEN** exit code 0

#### Scenario: Unhealthy

- **WHEN** the healthcheck subcommand runs and `/healthz` returns a non-200 status or connection error
- **THEN** exit code 1

### Requirement: Dependent state preservation

The system SHALL preserve the previous running state of dependent containers across healing operations.

#### Scenario: Stopped dependent

- **GIVEN** a dependent container was stopped before the master event
- **WHEN** eir heals (recreates) that dependent
- **THEN** it recreates the container but does NOT start it

### Requirement: Master stop is a no-op

The system SHALL NOT take action when a master container is stopped.

#### Scenario: Master stopped intentionally

- **WHEN** a master container is stopped (not restarted or recreated)
- **THEN** eir takes no action and waits for the master to return
