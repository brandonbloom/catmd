# Makefile for catmd

.PHONY: build test clean install lint fmt vet help dev

# Build the binary
build:
	go build -o catmd .

# Run all tests
test:
	./test.sh

# Clean build artifacts
clean:
	rm -f catmd
	go clean

# Install the binary to GOPATH/bin
install:
	go install .

# Run linter (requires golangci-lint)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Development checks (format, vet, lint, test)
dev: fmt vet lint test

# Update test expectations (use with caution)
update-tests:
	./test.sh --update

# Show available targets
help:
	@echo "Available targets:"
	@echo "  build        - Build the catmd binary"
	@echo "  test         - Run all tests"
	@echo "  clean        - Remove build artifacts"
	@echo "  install      - Install binary to GOPATH/bin"
	@echo "  lint         - Run golangci-lint (if installed)"
	@echo "  fmt          - Format Go code"
	@echo "  vet          - Run go vet"
	@echo "  dev          - Run format, vet, lint, and test"
	@echo "  update-tests - Update test expectations (use with caution)"
	@echo "  help         - Show this help message"

# Default target
all: dev
