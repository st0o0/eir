# syntax=docker/dockerfile:1@sha256:ecfaec9ed6d810b56388c508f4121597bfbba70d41a6dfeee4d8cad5f295fc32

FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 AS build
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
