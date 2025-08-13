.PHONY: help build test lint clean fmt lint-only

# Default target
help:
	@echo "Available targets:"
	@echo "  build    - Build the application"
	@echo "  test     - Run tests"
	@echo "  lint     - Format code and run golangci-lint"
	@echo "  fmt      - Format code using gofmt, goimports, and golangci-lint"
	@echo "  lint-only - Run golangci-lint without formatting"
	@echo "  clean    - Clean build artifacts"

# Build the application
build:
	go build -v -ldflags="-s -w" -o zigbee2mqtt-exporter ./cmd

# Run tests
test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Format code using gofmt, goimports, and golangci-lint formatters
fmt:
	go fmt ./...
	goimports -w .
	docker run --rm \
		-v "$(PWD):/app" \
		-w /app \
		golangci/golangci-lint:latest \
		golangci-lint run --fix

# Run golangci-lint using official container (formats first, then lints)
lint:
	make fmt
	docker run --rm \
		-v "$(PWD):/app" \
		-w /app \
		golangci/golangci-lint:latest \
		golangci-lint run

# Run only linting without formatting
lint-only:
	docker run --rm \
		-v "$(PWD):/app" \
		-w /app \
		golangci/golangci-lint:latest \
		golangci-lint run

# Clean build artifacts
clean:
	rm -f zigbee2mqtt-exporter coverage.txt
