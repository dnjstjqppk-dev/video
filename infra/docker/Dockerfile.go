# syntax=docker/dockerfile:1.7
# ==============================================================================
# Stage 1: Go Builder (Hardened Static Binary Compilation)
# Supports the API monolith (services/api) and every standalone worker module
# (services/<worker>) through the SVC_DIR / CMD_PATH build args.
# ==============================================================================
FROM golang:1.27-alpine AS builder

ARG SERVICE_NAME=api
ARG SVC_DIR=services/api
ARG CMD_PATH=./cmd/server
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

# Cache dependencies for the selected module only (go.sum optional: workers are stdlib-only)
COPY ${SVC_DIR}/go.mod ./services-tmp/
WORKDIR /src/services-tmp
RUN go mod download

# Copy full service tree and build
WORKDIR /src
COPY ${SVC_DIR}/ ./services-tmp/
WORKDIR /src/services-tmp

# Build static stripped binary with no CGO
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath \
    -ldflags="-s -w -X main.gitCommit=${GIT_COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -o /out/gvp-service ${CMD_PATH}

# ==============================================================================
# Stage 2: Minimal Non-Root Alpine Runtime (with FFmpeg for media-worker)
# ==============================================================================
FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates tzdata ffmpeg wget && \
    addgroup -g 10001 -S gvp && \
    adduser -u 10001 -S gvp -G gvp -h /app -s /sbin/nologin

WORKDIR /app

COPY --from=builder --chown=gvp:gvp /out/gvp-service /app/gvp-service

# Drop root privileges
USER 10001:10001

EXPOSE 8080 8081

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:${PORT:-8080}/healthz || exit 1

ENTRYPOINT ["/app/gvp-service"]
