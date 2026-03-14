# Project Guidelines

## Code Style
- Target Go 1.24 and keep changes consistent with the existing standard-library-first style in this repo.
- Keep public APIs small and explicit. This codebase commonly uses constructor helpers such as `NewPolicy`, `NewGroup`, and `NewPermission` that return pointers for domain types.
- Preserve the current testing style: `testify/assert`, `testify/mock`, and descriptive `Test...` names.

## Architecture
- `internal/authz` contains the core authorization domain: groups, permissions, policy evaluation, and the result model.
- `internal/authz/store` defines the generic `PolicyManager` contract and typed store errors. Storage implementations should conform to this package boundary instead of leaking backend-specific behavior.
- `internal/authz/store/postgres` contains the PostgreSQL-backed policy manager. It uses `pgx`, `slog`, optimistic concurrency via `version` columns, and SQL `MERGE` statements for set-style updates.
- `internal/shared` contains small reusable helpers such as the generic `Filter` function and test utilities. `internal/shared/testing` holds pgx mocks and testcontainer helpers.
- `cmd/authz` is the service entry point. Keep runtime wiring there and keep domain logic in `internal/...` packages.

## Build And Test
- Build everything with `go build -v ./...`.
- Run unit tests with `go test -v -short ./...`.
- Run integration tests with `go test -v -run Integration ./...`.
- Integration tests in `internal/authz/store/postgres` require Docker because they start PostgreSQL through `testcontainers-go` and initialize schema from `sql/authz_postgres.sql`.

## Conventions
- Return domain or store errors explicitly instead of swallowing invalid input. Examples in this repo include empty-user validation in `internal/authz` and `PolicyStoreError` wrappers in `internal/authz/store`.
- In the Postgres store, convert database-specific failures into `store.PolicyStoreError` values and log through the provided `slog.Logger` rather than returning raw pgx errors.
- For write operations on groups, preserve the optimistic concurrency pattern: read the current `version`, perform the change, then update or delete using the expected version and treat `RowsAffected() == 0` as a concurrency failure.
- Keep SQL schema assumptions aligned with `sql/authz_postgres.sql`; integration tests depend on that schema shape.
- When updating tests around PostgreSQL code, prefer the existing mock types in `internal/shared/testing` for unit coverage and the suite-based container tests for integration coverage.