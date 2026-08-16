# metrics

## Purpose

Prometheus metrics and HTTP health endpoint for observability of eir's event processing and healing operations.

## Requirements

### Requirement: Prometheus metrics endpoint

The system SHALL serve Prometheus metrics at `GET /metrics` on a configurable HTTP address (default `:9550`).

#### Scenario: Metrics available

- **WHEN** a Prometheus scraper requests `GET /metrics`
- **THEN** the response contains standard Go runtime metrics and eir-specific metrics in Prometheus exposition format

### Requirement: Event counter

The system SHALL count Docker events received via `eir_events_received_total` with an `action` label.

#### Scenario: Events counted by action

- **GIVEN** eir receives a `die` event for master `bifrost`
- **WHEN** the event is processed
- **THEN** `eir_events_received_total{action="die"}` increments by 1

### Requirement: Heal counter

The system SHALL count healing operations via `eir_heals_total` with `master` and `status` labels.

#### Scenario: Successful heal counted

- **GIVEN** eir heals dependents of master `bifrost` successfully
- **WHEN** the heal completes
- **THEN** `eir_heals_total{master="bifrost",status="success"}` increments by 1

#### Scenario: Failed heal counted

- **GIVEN** eir attempts to heal dependents of master `bifrost` and it fails
- **WHEN** the heal returns an error
- **THEN** `eir_heals_total{master="bifrost",status="fail"}` increments by 1

### Requirement: Heal duration histogram

The system SHALL observe healing duration via `eir_heal_duration_seconds` with a `master` label.

#### Scenario: Duration recorded

- **GIVEN** a heal operation for master `bifrost` takes 3.2 seconds
- **WHEN** the heal completes
- **THEN** `eir_heal_duration_seconds{master="bifrost"}` records 3.2

### Requirement: Health endpoint

The system SHALL serve a JSON health response at `GET /healthz` on the metrics server.

The handler SHALL ping the Docker daemon to verify reachability. The response SHALL include status, version, uptime, and the list of watched masters.

#### Scenario: Healthy response

- **WHEN** `GET /healthz` is requested and Docker is reachable
- **THEN** HTTP 200 with JSON body `{"status":"ok","version":"...","uptime":"...","masters":[...]}`

#### Scenario: Docker unreachable

- **WHEN** `GET /healthz` is requested and Docker is not reachable
- **THEN** HTTP 503 with JSON body `{"status":"error","error":"docker unreachable"}`
