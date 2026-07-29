## 1. Release Workflow

- [x] 1.1 Add `linux/arm/v7` to `platforms:` in `.github/workflows/release.yml`
- [x] 1.2 Verify the Dockerfile requires no changes for arm/v7 (Go cross-compiles natively, `scratch` is platform-agnostic)

## 2. Verify

- [x] 2.1 Build eir locally for `linux/arm/v7` with `docker buildx build --platform linux/arm/v7` to confirm it works
