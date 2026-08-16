## MODIFIED Requirements

### Requirement: Health check

The system SHALL provide a health check that verifies startup readiness, event-loop liveness, and Docker daemon reachability.

The main process SHALL write a heartbeat file at a fixed path on a periodic ticker (every 10 seconds) and after each processed event. The heartbeat file SHALL NOT be written until the event stream has been successfully established, providing explicit startup tracking.

The `healthcheck` subcommand SHALL check, in order:

1. Heartbeat file exists (startup readiness)
2. Heartbeat file age is less than 60 seconds (loop liveness)
3. Docker daemon responds to a ping within 5 seconds (daemon reachability)

#### Scenario: Startup not complete

- **WHEN** the health check runs and the heartbeat file does not exist
- **THEN** exit code 1 with message indicating startup is not complete

#### Scenario: Event loop stale

- **WHEN** the health check runs and the heartbeat file is older than 60 seconds
- **THEN** exit code 1 with message indicating a stale heartbeat

#### Scenario: Docker unreachable

- **WHEN** the health check runs and the heartbeat file is fresh but Docker does not respond within 5 seconds
- **THEN** exit code 1 with message indicating Docker is unreachable

#### Scenario: Fully healthy

- **WHEN** the health check runs and the heartbeat file exists, is fresh, and Docker responds
- **THEN** exit code 0
