## Context

Docker resolves `--network container:<name>` to `container:<full-ID>` at container creation time. After a master is recreated (`docker rm` + `docker run`), the dependent container's `HostConfig.NetworkMode` still contains the old master's full ID. The detector's `discoverByInspect` method compares this against the new master's name and new ID — neither matches, so the dependent is invisible.

The `lastKnownIDs` map in the detector already stores the previous container ID per master name. It is overwritten on line 72 of `detector.go` before `discoverDependents` is called, discarding the information needed to match the old reference.

## Goals / Non-Goals

**Goals:**

- Inspect-mode discovery finds dependents after a master recreate by matching against the previous master container ID
- Both restart and recreate scenarios are validated end-to-end against a real Docker daemon in CI
- The fix is backward-compatible — no new configuration, no behavioral change for restart case or label mode

**Non-Goals:**

- Tracking an unbounded history of previous IDs (only the immediately prior ID is needed)
- Changing label-mode discovery (already works because it matches on name)
- Adding new discovery modes

## Decisions

### Decision: Capture previous ID before overwrite

The fix captures `lastID` before the map assignment and passes it to `discoverByInspect` as a `previousID` parameter. The match condition becomes:

```
target == masterName || target == masterID || (previousID != "" && target == previousID)
```

**Alternative considered**: Store a list of all historical IDs per master. Rejected — only the immediately previous ID matters. After eir heals a recreate, the dependent is recreated with the new master ID, so older IDs are never referenced again.

**Alternative considered**: Resolve the ID stored in the dependent's `NetworkMode` back to a container name via `docker inspect`. Rejected — after `docker rm`, the old container no longer exists, so the inspect would fail.

### Decision: Raw Docker commands for e2e, not compose

The e2e test uses `docker run` / `docker rm` / `docker restart` directly. This gives explicit control over container IDs and avoids a docker-compose dependency in CI.

**Test containers**: nginx (master, serves on port 80), alpine with wget (dependent, verifies connectivity via `wget -qO /dev/null http://localhost:80`), eir:ci image (healer, watches master via docker.sock).

**Fast timings**: `EIR_STABILIZE_WAIT=2s`, `EIR_RETRY_BACKOFF=1s`, `EIR_MAX_RETRIES=2` to keep each scenario under 15 seconds.

### Decision: Separate test functions, shared cleanup

Each scenario (restart, recreate) runs as an independent function with full setup/teardown. A `trap` handler ensures cleanup on failure. Both scenarios run sequentially in one script invocation.

## Risks / Trade-offs

- **[Risk] Dependent's NetworkMode contains a name, not an ID** → Docker always resolves to the full ID at creation time, confirmed empirically. The `previousID` match is the correct path.
- **[Risk] Race between eir startup and master first-start event** → The e2e test starts eir before the master, then waits for eir to log the initial event. The detector classifies a first-seen start as `RecreateCase`, which is fine — it still discovers and heals dependents.
- **[Risk] e2e tests are slow** → Fast timings (2s stabilize, 1s backoff) keep each scenario under 15s. The CI job has a 15-minute timeout, and e2e only runs after unit/lint/build pass.
- **[Trade-off] Only testing inspect mode in e2e** → Label mode already works for recreates. The e2e focuses on the inspect-mode fix since that's the bug. Label-mode e2e can be added later if needed.
