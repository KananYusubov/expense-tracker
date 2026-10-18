# Personal Expense Tracker API (Initial Blueprint)

> **Notice:** This document serves as an initial planning blueprint and baseline architecture plan; implementation specifics, interfaces, and deliverables may evolve throughout active development.

## Execution & Setup

### Option 1: Docker (Recommended)

Run the full containerized environment with persistent storage in one command:

```
docker compose up --build
```

API binds to `http://localhost:8080` and stores SQLite data in `./data`.

### Option 2: Direct Terminal Execution

```
cd api
go test -v ./...
go run ./cmd/api
```

*Note: For Windows local development convenience, an `open-terminal.bat` script is included to automatically launch split terminal tabs for the Go runtime and Docker workflows.*

## Tech Stack & Architectural Decisions

* **Framework (`chi` v5):** Idiomatic, lightweight routing compatible with standard `net/http`, featuring built-in logging and panic recovery middlewares.
* **Database (`modernc.org/sqlite`):** Zero-setup persistence with a pure-Go driver that eliminates CGO dependencies for portable Docker builds.
* **Logging:** Structured request telemetry streamed to `stdout` following 12-Factor App principles.
* **Soft Deletes:** Deletions set a `deleted_at` timestamp so audit history is preserved while active endpoints filter for `deleted_at IS NULL`.

## Project Phases

### Phase 1 (Target Core)
* Core REST endpoints (`POST`, `GET`, `PATCH`, `DELETE`, `/expenses/summary`)
* Soft delete implementation (`deleted_at` tracking)
* Environment configuration (`PORT`, `DB_PATH`)
* Standardized request telemetry and structured logging
* Docker containerization and Docker Compose setup

### Phase 1.5 (Time Permitting)
* Input sanitization (trimming whitespace, stripping dangerous characters, string normalization)
* Unit and handler integration tests for validation rules
* Pagination (offset/limit) on the list endpoint

### Phase 2 (Future Scope)
* Granular differential PATCH updates (pointer fields)
* Dedicated category taxonomy tables for UI combobox selection
* Date-range and category query filtering
* Soft-delete restore endpoints
* Monetary migration from float to integer cents

## Edge Cases & Safeguards

* **Cents Precision:** Amounts strictly allow up to 2 decimal places; payloads exceeding 2 decimals return `400 Bad Request`.
* **Amount Ceiling:** Must be positive (`amount > 0`) with a sane upper limit to avoid overflow.
* **Soft Delete Isolation:** Querying or updating soft-deleted items returns `404`; `/expenses/summary` excludes deleted items.
* **Route Precedence:** `/expenses/summary` is registered before dynamic `/{id}` paths.