## MODIFIED Requirements

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
