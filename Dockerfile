# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o zigbee2mqtt-exporter ./cmd

# Final stage
FROM alpine:3.22.1

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/zigbee2mqtt-exporter .

# Expose port
EXPOSE 8087

# Run the binary
CMD ["./zigbee2mqtt-exporter"]
