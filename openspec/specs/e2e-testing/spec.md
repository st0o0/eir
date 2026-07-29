# E2E Testing

## Purpose

End-to-end test scenarios that verify eir's healing behavior against real Docker containers, covering restart and recreate cases with proper cleanup.

## Requirements

### Requirement: E2E restart scenario

The e2e test suite SHALL verify that eir heals dependent containers after a master container restart (same container ID).

#### Scenario: Restart healing restores connectivity

- **GIVEN** a master container running nginx on port 80
- **AND** a dependent container using `network_mode: container:<master>` that can reach localhost:80
- **AND** eir is running and watching the master
- **WHEN** the master container is restarted (`docker restart`)
- **THEN** eir detects the restart, heals the dependent, and the dependent can again reach localhost:80

### Requirement: E2E recreate scenario

The e2e test suite SHALL verify that eir heals dependent containers after a master container is removed and recreated (new container ID).

#### Scenario: Recreate healing restores connectivity

- **GIVEN** a master container running nginx on port 80
- **AND** a dependent container using `network_mode: container:<master>` that can reach localhost:80
- **AND** eir is running and watching the master
- **WHEN** the master container is removed (`docker rm -f`) and a new master is created with the same name
- **THEN** eir detects the recreate, recreates the dependent with the new master's network namespace, and the dependent can reach localhost:80

### Requirement: E2E cleanup on failure

The e2e test suite SHALL clean up all test containers regardless of test outcome.

#### Scenario: Cleanup after failure

- **WHEN** any e2e test scenario fails
- **THEN** all containers created by the test (master, dependent, eir) are removed before the script exits
