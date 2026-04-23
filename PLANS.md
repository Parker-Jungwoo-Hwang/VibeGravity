# PLANS.md

## Current work pack

Work Pack 01: Foundation and Repo Setup

## Goal

Fix the implementation base.  
When this pack is done: repo, core interfaces, schema skeleton, and dev loop are alive.

## Inputs

- docs: plans/07_workpack_foundation-and-repo-setup.md
- files: AGENTS.md, CLAUDE.md (already fixed)
- constraints: Go monorepo, PostgreSQL canonical, local embedding only, Codex-first reasoning

## Plan

1. Bootstrap Go monorepo (go mod init, directory structure, .gitignore)
2. Define core contracts (VibeGravityService interface, request/response DTOs)
3. Create storage baseline (Postgres migration skeleton for core tables)
4. Build runtime shell (api server + worker, both bootable separately)
5. Build config shell (dev config, CODEX_HOME, embedding endpoint as config values)
6. Add health check endpoint (GET /healthz)
7. Add doctor command (cmd/cli doctor — checks db, embedding, config)
8. Write baseline tests (app boots, db migrates, worker boots, health ok, config loads)

## Done when

- server and worker both boot locally
- health check returns ok
- migrations apply cleanly
- core interfaces are importable
- `go test ./...` passes
- `golangci-lint run` passes
- AGENTS.md, CLAUDE.md, PLANS.md are committed

## Risks

- cutting structure into microservices too early
- building routes before core contracts exist
- starting coding agents without instruction files
- not documenting the dev loop
