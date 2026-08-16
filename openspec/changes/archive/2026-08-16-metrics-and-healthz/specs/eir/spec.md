## MODIFIED Requirements

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
