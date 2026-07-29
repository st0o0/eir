## 1. Fix Detector

- [x] 1.1 In `Classify`, capture `lastID` before overwriting `lastKnownIDs` and pass it as `previousID` to `discoverDependents` / `discoverByInspect`
- [x] 1.2 In `discoverByInspect`, add `previousID` parameter and extend match condition: `target == masterName || target == masterID || (previousID != "" && target == previousID)`
- [x] 1.3 Add unit test `TestClassify_RecreateCase_DiscoversByPreviousID` — mock a dependent with `NetworkMode: container:<old-id>`, verify it is discovered when master starts with a new ID

## 2. E2E Test Script

- [x] 2.1 Create `tests/e2e/run.sh` with shared helpers: `cleanup()` with trap, `wait_for_log()` polling eir container logs, `check_connectivity()` running wget from dependent
- [x] 2.2 Implement restart scenario: start nginx master → start alpine dependent → verify connectivity → start eir → `docker restart` master → wait for "healed dependent" log → verify connectivity
- [x] 2.3 Implement recreate scenario: start nginx master → start alpine dependent → verify connectivity → start eir → `docker rm -f` master → `docker run` new master → wait for "healed dependent" log → verify connectivity on new dependent
- [x] 2.4 Verify both scenarios pass locally with `eir:ci` image

## 3. Verify CI

- [x] 3.1 Confirm `.github/workflows/ci.yml` e2e job works with the new `tests/e2e/run.sh` (no workflow changes expected)
