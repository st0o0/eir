# syntax=docker/dockerfile:1

FROM scratch
LABEL org.opencontainers.image.title="eir" \
      org.opencontainers.image.description="Docker container network healer — restores dependent containers when their master restarts or is recreated" \
      org.opencontainers.image.source="https://github.com/st0o0/eir" \
      org.opencontainers.image.documentation="https://github.com/st0o0/eir#readme" \
      org.opencontainers.image.licenses="MIT"
COPY eir /eir
COPY LICENSE NOTICE /

EXPOSE 9550

HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
  CMD ["/eir", "healthcheck"]

ENTRYPOINT ["/eir"]
