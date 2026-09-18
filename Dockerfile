# syntax=docker/dockerfile:1@sha256:ecfaec9ed6d810b56388c508f4121597bfbba70d41a6dfeee4d8cad5f295fc32

FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:4cb7ac979db5fcc41cae44b2227ba5ab8a51e8807f40d9ba4dee20a0ad960b5b AS build
ARG TARGETARCH
ARG VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /eir ./cmd/eir

FROM scratch
LABEL org.opencontainers.image.title="eir" \
      org.opencontainers.image.description="Docker container network healer — restores dependent containers when their master restarts or is recreated" \
      org.opencontainers.image.source="https://github.com/st0o0/eir" \
      org.opencontainers.image.documentation="https://github.com/st0o0/eir#readme" \
      org.opencontainers.image.licenses="MIT"
COPY --from=build /eir /eir
COPY LICENSE NOTICE /

EXPOSE 9550

HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
  CMD ["/eir", "healthcheck"]

ENTRYPOINT ["/eir"]
