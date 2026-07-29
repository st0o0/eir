#!/usr/bin/env bash
set -euo pipefail

# Container name prefix to avoid collisions.
P="eir-e2e"

MASTER="${P}-master"
DEP="${P}-dep"
EIR="${P}-eir"

PASS=0
FAIL=0

cleanup() {
  echo "--- cleanup ---"
  docker rm -f "$MASTER" "$DEP" "$EIR" 2>/dev/null || true
}
trap cleanup EXIT

# Wait until a string appears in a container's logs.
# Usage: wait_for_log <container> <pattern> [timeout_seconds]
wait_for_log() {
  local ctr="$1" pattern="$2" timeout="${3:-30}" elapsed=0
  while ! docker logs "$ctr" 2>&1 | grep -q "$pattern"; do
    sleep 1
    elapsed=$((elapsed + 1))
    if [ "$elapsed" -ge "$timeout" ]; then
      echo "TIMEOUT waiting for '$pattern' in $ctr logs"
      docker logs "$ctr" 2>&1 | tail -20
      return 1
    fi
  done
}

# Verify the dependent can reach nginx on localhost:80.
check_connectivity() {
  docker exec "$DEP" wget -qO /dev/null --timeout=5 http://localhost:80
}

start_eir() {
  docker run -d --name "$EIR" \
    -e EIR_MASTERS="$MASTER" \
    -e EIR_STABILIZE_WAIT=2s \
    -e EIR_RETRY_BACKOFF=1s \
    -e EIR_MAX_RETRIES=2 \
    -e EIR_LOG_LEVEL=debug \
    -v /var/run/docker.sock:/var/run/docker.sock:ro \
    eir:ci >/dev/null

  wait_for_log "$EIR" "eir starting"
}

# ---------- Scenario 1: Restart ----------

test_restart() {
  echo "=== Scenario: restart ==="
  cleanup

  echo "Starting eir first (so it sees master's initial start)..."
  start_eir

  echo "Starting master (nginx)..."
  docker run -d --name "$MASTER" nginx:alpine >/dev/null
  wait_for_log "$EIR" "received event"

  echo "Starting dependent..."
  docker run -d --name "$DEP" --network "container:${MASTER}" alpine sh -c "apk add --no-cache wget >/dev/null 2>&1; sleep 600" >/dev/null
  sleep 3

  echo "Verifying initial connectivity..."
  if ! check_connectivity; then
    echo "FAIL: initial connectivity check failed"
    return 1
  fi

  echo "Restarting master..."
  docker restart "$MASTER" >/dev/null

  echo "Waiting for eir to heal..."
  if ! wait_for_log "$EIR" "healed dependent" 30; then
    echo "FAIL: eir did not heal dependent after restart"
    docker logs "$EIR" 2>&1 | tail -30
    return 1
  fi

  sleep 2
  echo "Verifying connectivity after heal..."
  if ! check_connectivity; then
    echo "FAIL: connectivity check failed after restart heal"
    return 1
  fi

  echo "PASS: restart scenario"
}

# ---------- Scenario 2: Recreate ----------

test_recreate() {
  echo "=== Scenario: recreate ==="
  cleanup

  echo "Starting eir first (so it sees master's initial start)..."
  start_eir

  echo "Starting master (nginx)..."
  docker run -d --name "$MASTER" nginx:alpine >/dev/null
  wait_for_log "$EIR" "received event"

  echo "Starting dependent..."
  docker run -d --name "$DEP" --network "container:${MASTER}" alpine sh -c "apk add --no-cache wget >/dev/null 2>&1; sleep 600" >/dev/null
  sleep 3

  echo "Verifying initial connectivity..."
  if ! check_connectivity; then
    echo "FAIL: initial connectivity check failed"
    return 1
  fi

  echo "Removing master..."
  docker rm -f "$MASTER" >/dev/null

  echo "Recreating master with new ID..."
  docker run -d --name "$MASTER" nginx:alpine >/dev/null

  echo "Waiting for eir to heal..."
  if ! wait_for_log "$EIR" "healed dependent" 30; then
    echo "FAIL: eir did not heal dependent after recreate"
    docker logs "$EIR" 2>&1 | tail -30
    return 1
  fi

  sleep 2
  echo "Verifying connectivity after heal..."
  if ! check_connectivity; then
    echo "FAIL: connectivity check failed after recreate heal"
    return 1
  fi

  echo "PASS: recreate scenario"
}

# ---------- Main ----------

echo "eir e2e tests"
echo "============="

if test_restart; then
  PASS=$((PASS + 1))
else
  FAIL=$((FAIL + 1))
fi

if test_recreate; then
  PASS=$((PASS + 1))
else
  FAIL=$((FAIL + 1))
fi

echo ""
echo "Results: ${PASS} passed, ${FAIL} failed"

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
