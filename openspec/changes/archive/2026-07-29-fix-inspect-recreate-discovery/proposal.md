## Why

Docker resolves `--network container:<name>` to `container:<full-container-ID>` at creation time. When a master container is recreated (new ID), the dependent's stored `NetworkMode` still references the old master ID. The detector's inspect-mode discovery only matches against the current master name and new master ID, so it never finds the dependent — healing silently fails for the recreate case under inspect mode.

This is a correctness bug: the most common deployment pattern (`network_mode: "container:<name>"` in docker-compose) breaks on master recreates unless users opt into label mode.

## What Changes

- **Fix inspect-mode dependent discovery**: Pass the previous master container ID into `discoverByInspect` and match dependents whose `NetworkMode` references the old ID. The `lastKnownIDs` map already tracks this — the old ID just needs to be captured before it is overwritten and forwarded to the match logic.
- **Update event-detection spec**: Add a requirement covering inspect-mode discovery after a recreate (matching against the previous container ID).
- **Add e2e tests**: Create `tests/e2e/run.sh` (already wired in CI) with two scenarios — restart and recreate — using raw Docker commands against a real Docker daemon.

## Capabilities

### New Capabilities

- `e2e-testing`: End-to-end test infrastructure for validating healing scenarios against a real Docker daemon.

### Modified Capabilities

- `event-detection`: Inspect-mode dependent discovery must also match against the previous master container ID after a recreate.

## Impact

- `internal/detector/detector.go` — signature change for `discoverByInspect`, updated match logic
- `internal/detector/detector_test.go` — new test case for recreate discovery with old ID
- `tests/e2e/run.sh` — new file (restart + recreate scenarios)
- `openspec/specs/event-detection/spec.md` — updated requirement
