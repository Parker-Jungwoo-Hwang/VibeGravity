.PHONY: build test eval integration-postgres migration-postgres lint check-headers check-shell vuln verify-modules release-gate release-checksums sbom clean dev-server dev-worker setup

GOLANGCI_LINT ?= $(shell command -v golangci-lint 2>/dev/null || printf "%s/bin/golangci-lint" "$$(go env GOPATH)")
GOLANGCI_LINT_VERSION ?= v2.5.0
GOVULNCHECK ?= $(shell command -v govulncheck 2>/dev/null || printf "%s/bin/govulncheck" "$$(go env GOPATH)")
GOVULNCHECK_VERSION ?= v1.1.4
SHELLCHECK ?= shellcheck
SHELL_SCRIPTS := $(shell find .agents -type f -name '*.sh' | sort)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || printf dev)
CHECKSUM_DIR ?= dist
CHECKSUM_FILE ?= $(CHECKSUM_DIR)/checksums.txt
SBOM_FILE ?= $(CHECKSUM_DIR)/sbom.cdx.json

# Build all binaries
build:
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker
	go build -ldflags "-X main.version=$(VERSION)" -o bin/vibegravity ./cmd/vibegravity

# Run all tests
test:
	go test -v ./...

# Run deterministic golden quality evals
eval:
	go run ./cmd/vibegravity eval golden --path tests/golden/replay_eval.json
	go run ./cmd/vibegravity eval demo

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

# Run opt-in migration apply and rollback smoke against a separate empty DB.
migration-postgres:
	@if [ -z "$$VIBEGRAVITY_MIGRATION_TEST_DB_URL" ]; then \
		printf "%s\n" "Skipping migration smoke: VIBEGRAVITY_MIGRATION_TEST_DB_URL is not set."; \
		printf "%s\n" "Prepare an empty scratch DB, export VIBEGRAVITY_MIGRATION_TEST_DB_URL, then rerun: make migration-postgres"; \
	else \
		printf "%s\n" "Running migration apply/rollback smoke against VIBEGRAVITY_MIGRATION_TEST_DB_URL."; \
		go test -v -count=1 ./tests -run TestMigrationsApplyOnEmptyDatabaseAndRollbackSmoke; \
	fi

# Run linter
lint:
	$(GOLANGCI_LINT) run

# Run Go vulnerability checks
vuln:
	$(GOVULNCHECK) ./...

# Verify downloaded module content against go.sum.
verify-modules:
	go mod verify

# Run the local release gate for a release candidate or private validation drop.
release-gate:
	go test -count=1 ./...
	$(MAKE) eval
	$(MAKE) integration-postgres
	$(MAKE) migration-postgres
	$(MAKE) lint
	$(MAKE) check-headers
	$(MAKE) check-shell
	git diff --check
	go build ./cmd/server ./cmd/worker ./cmd/vibegravity
	go mod verify
	$(GOVULNCHECK) ./...

# Generate SHA-256 checksums for staged release artifacts.
release-checksums:
	@if [ ! -d "$(CHECKSUM_DIR)" ]; then \
		printf "%s\n" "Missing $(CHECKSUM_DIR). Stage release artifacts first, then rerun."; \
		exit 1; \
	fi
	@find "$(CHECKSUM_DIR)" -type f ! -name "$$(basename "$(CHECKSUM_FILE)")" ! -name "$$(basename "$(SBOM_FILE)")" -print | sort | while IFS= read -r file; do \
		shasum -a 256 "$$file"; \
	done > "$(CHECKSUM_FILE)"
	@test -s "$(CHECKSUM_FILE)" || { printf "%s\n" "No release artifacts found in $(CHECKSUM_DIR)."; rm -f "$(CHECKSUM_FILE)"; exit 1; }
	@printf "%s\n" "Wrote $(CHECKSUM_FILE)"

# Generate an optional CycloneDX SBOM with Syft.
sbom:
	@if ! command -v syft >/dev/null 2>&1; then \
		printf "%s\n" "Missing syft. Install syft before generating $(SBOM_FILE)."; \
		exit 1; \
	fi
	@mkdir -p "$(CHECKSUM_DIR)"
	syft dir:. -o cyclonedx-json > "$(SBOM_FILE)"

# Check source file headers
check-headers:
	go run ./tools/headercheck

# Check repo-local agent shell scripts.
check-shell:
	@for script in $(SHELL_SCRIPTS); do \
		case "$$(head -n 1 "$$script")" in \
			*bash*) bash -n "$$script" ;; \
			*) sh -n "$$script" ;; \
		esac; \
	done
	@command -v "$(SHELLCHECK)" >/dev/null 2>&1 || { printf "%s\n" "Missing shellcheck. Install shellcheck or set SHELLCHECK=/path/to/shellcheck."; exit 1; }
	$(SHELLCHECK) -x $(SHELL_SCRIPTS)

# Clean local build artifacts. This removes bin/.
clean:
	rm -rf bin/

# Run dev server
dev-server:
	go run ./cmd/server

# Run dev worker
dev-worker:
	go run ./cmd/worker

# Setup environment (useful for new devs).
# This downloads Go modules and installs pinned external Go tools.
setup:
	go mod download
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
