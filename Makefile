.PHONY: build test eval integration-postgres lint check-headers clean dev-server dev-worker setup

GOLANGCI_LINT ?= $(shell command -v golangci-lint 2>/dev/null || printf "%s/bin/golangci-lint" "$$(go env GOPATH)")

# Build all binaries
build:
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker
	go build -o bin/cli ./cmd/cli

# Run all tests
test:
	go test -v ./...

# Run deterministic golden quality evals
eval:
	go run ./cmd/cli eval golden --path tests/golden/replay_eval.json
	go run ./cmd/cli eval demo

# Run opt-in live PostgreSQL integration checks.
# Keeps the default local gate deterministic: `make test` still works without DB.
integration-postgres:
	@if [ -z "$$VIBEGRAVITY_DB_URL" ]; then \
		printf "%s\n" "Skipping live PostgreSQL integration gate: VIBEGRAVITY_DB_URL is not set."; \
		printf "%s\n" "Prepare a migrated scratch DB, export VIBEGRAVITY_DB_URL, then rerun: make integration-postgres"; \
	else \
		printf "%s\n" "Running live PostgreSQL integration gate against VIBEGRAVITY_DB_URL."; \
		go test -v -count=1 ./internal/store/postgres ./internal/kernel ./tests; \
	fi

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
