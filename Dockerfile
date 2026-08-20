# Build stage
# Use golang:1.25.4-alpine which uses Alpine 3.22.2 to avoid qemu emulation issues
FROM golang:1.27.0-alpine@sha256:4c9fe60190a2a3350ddc51de80d0224b8a6698d12bdfc999fee45ea9d6c46dbc AS builder

WORKDIR /app

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with version information
# Accept build args for version info, fall back to git describe if not provided
ARG VERSION
ARG COMMIT
ARG BUILD_DATE

RUN VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")} && \
    COMMIT=${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")} && \
    BUILD_DATE=${BUILD_DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")} && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-s -w \
        -X github.com/d0ugal/zigbee2mqtt-exporter/internal/version.Version=$VERSION \
        -X github.com/d0ugal/zigbee2mqtt-exporter/internal/version.Commit=$COMMIT \
        -X github.com/d0ugal/zigbee2mqtt-exporter/internal/version.BuildDate=$BUILD_DATE" \
    -o zigbee2mqtt-exporter ./cmd

# Final stage
FROM alpine:3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

RUN apk --no-cache add ca-certificates

# Setup an unprivileged user
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser

WORKDIR /app
RUN chown appuser:appgroup /app

USER appuser

# Copy the binary from builder stage
COPY --from=builder --chown=appuser:appuser /app/zigbee2mqtt-exporter .

# Expose port
EXPOSE 8087

# Run the binary
CMD ["/app/zigbee2mqtt-exporter"]
