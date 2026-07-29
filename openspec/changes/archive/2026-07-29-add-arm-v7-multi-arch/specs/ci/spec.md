## ADDED Requirements

### Requirement: ARM v7 platform support

The release workflow SHALL produce Docker images for `linux/arm/v7` in addition to the existing `linux/amd64` and `linux/arm64` platforms.

#### Scenario: Multi-arch manifest includes arm/v7

- **WHEN** a release is created
- **THEN** the published image manifest at `ghcr.io/st0o0/eir:<version>` includes `linux/amd64`, `linux/arm64`, and `linux/arm/v7` platforms
