## Why

eir currently builds for `linux/amd64` and `linux/arm64`. Users running Docker on 32-bit ARM devices (Raspberry Pi 2/3, older Synology NAS) cannot pull the image. Adding `linux/arm/v7` covers the remaining common Docker platform and aligns with the homelab/sidecar use case.

## What Changes

- Add `linux/arm/v7` to the release workflow's `platforms:` list
- Switch from shared GHA cache to per-platform cache scopes to prevent cross-platform cache eviction (relevant now with 3 platforms)
- Dev-build remains amd64-only (no change)

## Capabilities

### New Capabilities

None — this is a CI/workflow change only.

### Modified Capabilities

None — no spec-level behavior changes. eir's runtime behavior is identical across platforms.

## Impact

- `.github/workflows/release.yml` — platform list and cache configuration
- No application code changes
- No Dockerfile changes (Go cross-compiles natively with `CGO_ENABLED=0`, `scratch` base is platform-agnostic)
