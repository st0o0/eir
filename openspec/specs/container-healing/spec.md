# Container Healing

## Purpose

Executes recovery operations on dependent containers after a master container event. Handles both restart and recreate cases, with stabilization delay and retry logic.

## Requirements

### Requirement: Stabilization wait

The system SHALL wait for a configurable duration after a master start event before healing dependents, giving the master time to fully initialize.

#### Scenario: Default wait

- **GIVEN** `EIR_STABILIZE_WAIT=15s`
- **WHEN** a master start event is detected
- **THEN** eir waits 15 seconds before healing dependents

#### Scenario: Context cancellation during wait

- **WHEN** eir is shut down during the stabilization wait
- **THEN** the wait is cancelled and no healing occurs

### Requirement: Restart healing

When a master is restarted (same container ID), the system SHALL restart each dependent container.

#### Scenario: Dependent restart

- **GIVEN** master event classified as `RestartCase`
- **AND** `app1` is a dependent
- **WHEN** healing runs
- **THEN** `docker restart app1` is executed

### Requirement: Recreate healing

When a master is recreated (new container ID), the system SHALL recreate each dependent container with the updated `NetworkMode` pointing to the new master container ID.

#### Scenario: Running dependent recreated

- **GIVEN** master event classified as `RecreateCase` with new master ID `def456`
- **AND** dependent `app1` was running
- **WHEN** healing runs
- **THEN** eir stops `app1`, removes it, creates a new container with `NetworkMode: container:def456` and the same config, then starts it

#### Scenario: Stopped dependent recreated

- **GIVEN** master event classified as `RecreateCase`
- **AND** dependent `app1` was stopped
- **WHEN** healing runs
- **THEN** eir stops, removes, and creates `app1` with the new NetworkMode but does NOT start it

### Requirement: Retry with exponential backoff

The system SHALL retry failed healing operations with exponential backoff up to a configurable maximum number of attempts.

#### Scenario: Retry on failure

- **GIVEN** `EIR_MAX_RETRIES=3` and `EIR_RETRY_BACKOFF=5s`
- **WHEN** a heal operation fails
- **THEN** it retries after 5s, 10s, 20s, then gives up

#### Scenario: All retries exhausted

- **GIVEN** `EIR_MAX_RETRIES=3`
- **WHEN** all 3 retry attempts fail
- **THEN** the error is logged and eir continues with the next dependent

### Requirement: No dependents is a no-op

The system SHALL handle the case where a master has no dependents gracefully.

#### Scenario: No dependents found

- **WHEN** a master event triggers healing but no dependents are discovered
- **THEN** eir logs this fact and takes no further action
