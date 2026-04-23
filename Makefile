.PHONY: build test lint check-headers clean dev-server dev-worker setup

GOLANGCI_LINT ?= $(shell command -v golangci-lint 2>/dev/null || printf "%s/bin/golangci-lint" "$$(go env GOPATH)")

# Build all binaries
build:
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker
	go build -o bin/cli ./cmd/cli

# Run all tests
test:
	go test -v ./...

# Run linter
lint:
	$(GOLANGCI_LINT) run

# Check source file headers
check-headers:
	go run ./tools/headercheck

# Clean build artifacts
clean:
	rm -rf bin/

# Run dev server
dev-server:
	go run ./cmd/server

# Run dev worker
dev-worker:
	go run ./cmd/worker

# Setup environment (useful for new devs)
setup:
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
