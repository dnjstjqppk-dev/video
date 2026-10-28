# syntax=docker/dockerfile:1.7
# ==============================================================================
# Stage 1: Go Builder (Hardened Static Binary Compilation)
# ==============================================================================
FROM golang:1.27-alpine3.20 AS builder

ARG SERVICE_NAME=api
ARG CMD_PATH=./cmd/server
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

# Cache dependencies
COPY services/api/go.mod services/api/go.sum* ./services/api/
WORKDIR /src/services/api
RUN go mod download && go mod verify

# Copy full Go service tree
COPY services/api/ ./

# Build static stripped binary with no CGO
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath \
    -ldflags="-s -w -X main.gitCommit=${GIT_COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -o /out/gvp-service ${CMD_PATH}

# ==============================================================================
# Stage 2: Minimal Non-Root Distroless / Alpine Runtime (with FFmpeg for media-worker)
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
  CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["/app/gvp-service"]
