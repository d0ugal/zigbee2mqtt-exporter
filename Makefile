.PHONY: help build test lint clean fmt lint-only

# Default target
help:
	@echo "Available targets:"
	@echo "  build    - Build the application"
	@echo "  test     - Run tests"
	@echo "  lint     - Format code and run golangci-lint"
	@echo "  fmt      - Format code using golangci-lint"
	@echo "  lint-only - Run golangci-lint without formatting"
	@echo "  clean    - Clean build artifacts"

# Build the application
build:
	@echo "Building zigbee2mqtt-exporter..."
	@VERSION=$$(git describe --tags --always --dirty 2>/dev/null || echo "dev") && \
	COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") && \
	BUILD_DATE=$$(date -u +"%Y-%m-%dT%H:%M:%SZ") && \
	go build -v -ldflags="-s -w \
		-X github.com/d0ugal/zigbee2mqtt-exporter/internal/version.Version=$$VERSION \
		-X github.com/d0ugal/zigbee2mqtt-exporter/internal/version.Commit=$$COMMIT \
		-X github.com/d0ugal/zigbee2mqtt-exporter/internal/version.BuildDate=$$BUILD_DATE" \
		-o zigbee2mqtt-exporter ./cmd

# Run tests
test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./internal/... || true

# Format code using golangci-lint formatters (faster than separate tools)
fmt:
	docker run --rm \
		-v "$(PWD):/app" \
		-v "$(HOME)/.cache:/root/.cache" \
		-w /app \
		golangci/golangci-lint:latest \
		golangci-lint run --fix

# Run golangci-lint (formats first, then lints)
lint:
	docker run --rm \
		-v "$(PWD):/app" \
		-v "$(HOME)/.cache:/root/.cache" \
		-w /app \
		golangci/golangci-lint:latest \
		golangci-lint run --fix

# Run only linting without formatting
lint-only:
	docker run --rm \
		-v "$(PWD):/app" \
		-v "$(HOME)/.cache:/root/.cache" \
		-w /app \
		golangci/golangci-lint:latest \
		golangci-lint run

# Clean build artifacts
clean:
	rm -f zigbee2mqtt-exporter coverage.txt
