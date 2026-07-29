## Context

The release workflow (`release.yml`) currently builds for `linux/amd64,linux/arm64` in a single `docker/build-push-action` step with QEMU emulation and a shared GHA cache. The Dockerfile uses `FROM scratch` and copies a statically-linked Go binary — QEMU only emulates the `COPY` instruction, making cross-platform builds fast.

## Goals / Non-Goals

**Goals:**

- Support `linux/arm/v7` in release builds
- Prevent cache eviction between platforms with per-platform cache scopes

**Non-Goals:**

- Adding arm/v7 to dev-build (amd64-only is sufficient for PR testing)
- Splitting into separate per-platform build jobs (the njord pattern — unnecessary for Go + scratch)
- Adding linux/386 or other niche platforms

## Decisions

### Decision: Keep single build-push-action with QEMU

The njord project uses a matrix of separate `dotnet publish` jobs per platform, then merges manifests with `imagetools create`. This is necessary for .NET because QEMU-emulated compilation is extremely slow.

For eir, Go cross-compiles natively (`CGO_ENABLED=0` + `GOARCH` set by Buildx), and the `scratch` base image has no package installation. QEMU only runs for the final `COPY` instruction. Adding arm/v7 to the existing `platforms:` line is sufficient.

### Decision: Per-platform cache scopes

With 3 platforms sharing one GHA cache key, platforms can evict each other's layers. Switching to per-platform scopes (`scope=amd64`, `scope=arm64`, `scope=armv7`) prevents this. This follows the pattern used in the njord release workflow.

The `cache-from` needs all three scopes listed so each platform build can read from its own cache. The `cache-to` writes to a single scope matching the current build — Buildx handles this automatically when `cache-to` contains `scope=build` but since we're using a single build step for all platforms, we use a combined `cache-from` with all scopes and a shared `cache-to`.

Actually, since `build-push-action` builds all platforms in one invocation, the cache scope cannot be per-platform without splitting into separate jobs. The simpler approach: keep the shared cache (`type=gha`) as-is. With Go's fast compilation and the minimal `scratch` layer, cache efficiency is not a bottleneck.

**Revised decision**: Keep shared GHA cache. The added complexity of splitting into per-platform jobs isn't justified for Go + scratch builds.

## Risks / Trade-offs

- **[Risk] arm/v7 build adds ~1-2 minutes to release time** → Acceptable for a release-only step that runs infrequently
- **[Risk] QEMU arm/v7 emulation issues** → Minimal risk since QEMU only runs `COPY` on `scratch`, not actual compilation
