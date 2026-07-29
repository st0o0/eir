# Event Detection

## Purpose

Subscribes to the Docker event stream, filters for master container lifecycle events, classifies them as restart or recreate, and discovers which containers depend on each master.

## Requirements

### Requirement: Event stream subscription

The system SHALL subscribe to the Docker event stream filtered to configured master containers and the event types `die` and `start`.

#### Scenario: Filtered subscription

- **GIVEN** `EIR_MASTERS=bifrost`
- **WHEN** eir subscribes to Docker events
- **THEN** only events for container `bifrost` with action `die` or `start` are received

### Requirement: Restart classification

The system SHALL classify a master event as a **restart** when the container ID remains the same across a die→start cycle.

#### Scenario: Same container ID

- **GIVEN** master `bifrost` with container ID `abc123`
- **WHEN** a `die` event for `abc123` is followed by a `start` event for `abc123`
- **THEN** the event is classified as `RestartCase`

### Requirement: Recreate classification

The system SHALL classify a master event as a **recreate** when a new container ID appears for the same master name.

#### Scenario: New container ID

- **GIVEN** master `bifrost` was previously seen with container ID `abc123`
- **WHEN** a `start` event arrives for `bifrost` with container ID `def456`
- **THEN** the event is classified as `RecreateCase`

### Requirement: First-seen start event

The system SHALL ignore the first `start` event for a master that has no previously known container ID (initial startup, not a recreate).

#### Scenario: Initial start

- **WHEN** eir starts and receives a `start` event for a master it has never seen
- **THEN** it records the container ID but does not trigger healing

### Requirement: Dependent discovery via inspect mode

In `inspect` mode, the system SHALL discover dependents by inspecting all containers and matching those whose `NetworkMode` equals `container:<master-container-id>`, `container:<master-name>`, or `container:<previous-master-container-id>` (the container ID from before a recreate).

#### Scenario: Inspect discovery

- **GIVEN** `EIR_DISCOVERY_MODE=inspect`
- **AND** containers `app1` and `app2` have `NetworkMode: container:<bifrost-id>`
- **WHEN** a master event occurs for `bifrost`
- **THEN** `app1` and `app2` are identified as dependents

#### Scenario: Inspect discovery after recreate

- **GIVEN** `EIR_DISCOVERY_MODE=inspect`
- **AND** master `bifrost` was previously known with container ID `old-abc123`
- **AND** a new `bifrost` container starts with ID `new-def456`
- **AND** dependent `app1` has `NetworkMode: container:old-abc123`
- **WHEN** the detector discovers dependents
- **THEN** `app1` is identified as a dependent (matched via previous master ID)

### Requirement: Dependent discovery via label mode

In `label` mode, the system SHALL discover dependents by filtering containers with the label `eir.master=<master-name>`.

#### Scenario: Label discovery

- **GIVEN** `EIR_DISCOVERY_MODE=label`
- **AND** container `app1` has label `eir.master=bifrost`
- **WHEN** a master event occurs for `bifrost`
- **THEN** `app1` is identified as a dependent
