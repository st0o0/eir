# Changelog

## [0.1.4](https://github.com/st0o0/eir/compare/v0.1.3...v0.1.4) (2026-11-20)


### Features

* extend docker preset for base image digest pinning + automerge ([5e6bc88](https://github.com/st0o0/eir/commit/5e6bc88d23b343cdcc55377448e06e10e5fe701c))
* migrate to modular build and docker workflows ([fda4539](https://github.com/st0o0/eir/commit/fda453914bf1e746ff64f63f253fb170e4df819f))
* migrate to multi-stage Dockerfile ([eae4a17](https://github.com/st0o0/eir/commit/eae4a17fa5b3a50ddc156e759fa17c6351299528))
* **security:** add source scan on PR/push, registry scan after release ([ede46ca](https://github.com/st0o0/eir/commit/ede46ca859441a8cbd3cdaa55a5c1fc8fdcc07cb))


### Bug Fixes

* add id-token permission for cosign signing in dev builds ([84a2455](https://github.com/st0o0/eir/commit/84a2455933c107079b69a4e5e0eab60d875c7813))
* grant contents:read and drop paths filter in dev-build docker job ([c880883](https://github.com/st0o0/eir/commit/c88088330fbba7f886a7dbcb3bb182545d842f59))
* grant pull-requests/issues write permission to labeler and label-sync callers ([2b164ee](https://github.com/st0o0/eir/commit/2b164ee95dce7d35db24a5b285e27670ba99c030))
* **security:** add packages:read permission for reusable workflow compatibility ([77b9478](https://github.com/st0o0/eir/commit/77b947816e3cf8c177da7d08adfe6619a2fce622))
* **security:** add trivy.yaml with DS-0002 skip and wire trivy-config ([806d327](https://github.com/st0o0/eir/commit/806d327015f63b59d35c32478304154ce6490484))
* **security:** use .trivyignore for DS-0002 skip instead of trivy.yaml ([94d4194](https://github.com/st0o0/eir/commit/94d41943683bdc078245c988db275277dae59349))


### Documentation

* align README badges ([1698881](https://github.com/st0o0/eir/commit/1698881c91b05925c4c10177462a0f5d85e0fc38))
* remove em dashes and simplify wording ([189de82](https://github.com/st0o0/eir/commit/189de82509117be24188fb4a73ec7c575a25e968))


### Refactoring

* **security:** remove unused trivy.yaml placeholder ([afd9a77](https://github.com/st0o0/eir/commit/afd9a772233fb3e472edcc7832f8ba2dbae3de6f))
* **security:** simplify to fs scan on PR/push and local image scan on schedule ([82fc17c](https://github.com/st0o0/eir/commit/82fc17c1b99246e5ebdbb70a9df8af2edfc7b5f8))

## [0.1.3](https://github.com/st0o0/eir/compare/v0.1.2...v0.1.3) (2026-09-12)


### Features

* decouple release-please from build workflow ([b5546d9](https://github.com/st0o0/eir/commit/b5546d90c9d519f484fe8b692b94d8c09726cb76))


### Refactoring

* migrate to shared reusable workflows ([eca5e94](https://github.com/st0o0/eir/commit/eca5e94558f857bca5a70d42b42331a4308eb9e7))
* rename CI jobs for cleaner GitHub check names ([038a7b4](https://github.com/st0o0/eir/commit/038a7b48be7cd42538657fc81dc0ffdefe11a06d))


### Dependencies

* bump hadolint/hadolint-action from 3.4.0 to 3.5.0 in the actions-all group ([#10](https://github.com/st0o0/eir/issues/10)) ([ec0ae2a](https://github.com/st0o0/eir/commit/ec0ae2af52a4a46c2a20a6723635ceaedca611ea))
* bump hadolint/hadolint-action in the actions-all group ([94d34a0](https://github.com/st0o0/eir/commit/94d34a062eccbfbe4c3fc19cea7a6fcbbfdf7547))

## [0.1.2](https://github.com/st0o0/eir/compare/v0.1.1...v0.1.2) (2026-08-16)


### Features

* add Prometheus metrics and HTTP-based health check ([778b4b5](https://github.com/st0o0/eir/commit/778b4b5298148009b8cd443f28d12f6d74fce8d0))
* Use new Go collector ([40507e9](https://github.com/st0o0/eir/commit/40507e9c3c200559150f837a784ae6834de59b52))

## [0.1.1](https://github.com/st0o0/eir/compare/v0.1.0...v0.1.1) (2026-07-30)


### Bug Fixes

* **healer:** clear port config and stop re-running teardown on retry ([e7169a1](https://github.com/st0o0/eir/commit/e7169a17d988484537ceacaf147ab14946665982))

## [0.1.0](https://github.com/st0o0/eir/compare/v0.1.0...v0.1.0) (2026-07-29)


### Features

* **ci:** Add ARM v7 platform support to release builds ([e5d030d](https://github.com/st0o0/eir/commit/e5d030d4b810f0372a62ed1de8edd284965d76a9))
* initial project scaffold ([b505d3f](https://github.com/st0o0/eir/commit/b505d3f4ece96ef0affe02b4bbe101e70a4ea4cf))
* introduce OpenSpec commands and skills ([27fe09e](https://github.com/st0o0/eir/commit/27fe09ef989fd21c6bcd9161345b1d8d47cd668e))


### Bug Fixes

* inspect-mode dependent discovery after master recreate ([321e367](https://github.com/st0o0/eir/commit/321e367586a2334a8ed1918010bd16678a5d3c58))
* upgrade docker/docker v27.5.1 → v28.5.2 to resolve security alerts ([ab86a50](https://github.com/st0o0/eir/commit/ab86a502780099eb0eb500282a855227634d7100))
