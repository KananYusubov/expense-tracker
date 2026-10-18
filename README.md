# Personal Expense Tracker API

A lightweight, robust REST API for tracking personal expenses built with Go, Chi router, and pure-Go SQLite persistence.

---

## 1. How to Run the Project

### Option A: Docker Compose (Recommended)

> **Prerequisites:**
> - Ensure **Docker Desktop** is open and running on your system before executing the command.
> - ⏱️ **First run build time:** Running `docker compose up --build` for the first time takes **approximately 2 minutes** to download base images and dependencies. Subsequent runs start in seconds.

```bash
docker compose up --build
```
API runs at `http://localhost:8080`. SQLite database and JSON logs persist in `./data/`.

### Option B: Local Go Execution

```bash
cd expense-api
go run ./api
```


---

## 2. Technology Choices & Rationale

- **Framework (`chi` v5):** Idiomatic, lightweight, and fully compatible with Go's standard `net/http`. Provides composable middleware without heavy framework overhead.
- **Database (`modernc.org/sqlite`):** Pure-Go SQLite requiring zero CGO dependencies for portable, minimal Alpine Docker images. All queries are parameterized to prevent SQL injection.
- **Structured JSON Logger (`internal/logger`):** Custom middleware recording request telemetry (`time`, `level`, `method`, `path`, `status`, `duration`, `errorMessage`) to `./data/app.log` and `stdout` without boilerplate in handlers.
- **Input Sanitization (`internal/sanitize`):** Centralized helper that trims inputs, normalizes categories, and escapes HTML to prevent stored XSS.

---

## 3. Assumptions Made

- **Route Precedence:** `/expenses/summary` is registered before dynamic `/{id:[0-9]+}` routes to avoid URL collisions.
- **Summary Format:** Grouped by category with overall metrics: `{"total_amount": ..., "total_expenses": ..., "total_categories": ..., "categories": {"food": ...}}`.
- **Soft Deletes:** Deletions set `deleted_at`, preserving audit history while excluding records from active endpoints and summary calculations.
- **Precision & Ceiling:** Amount must be positive (`> 0`) with up to 2 decimal places and capped by `MAX_EXPENSE_AMOUNT` to prevent numerical overflow.
- **Single User Scope:** Authentication is omitted per requirements; single-tenant usage is assumed.

---

## 4. What I Would Improve With More Time

- **Automated Testing:** Implement comprehensive unit and integration test suites covering HTTP handlers, edge cases, and storage layers.
- **Category Taxonomy:** Dedicated database tables for categories to support predefined taxonomies, metadata, and UI combobox selection.
- **Pagination & Filtering:** Add `limit`/`offset` pagination and query filters (date range `from`/`to`, category) to `GET /expenses`.
- **Monetary Precision:** Migrate financial amounts from `float64` to integer cents (`int64`) to eliminate potential floating-point rounding quirks.
- **Soft Delete Restore:** Add an endpoint to inspect audit history and recover soft-deleted expenses.