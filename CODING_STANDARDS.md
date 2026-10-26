# Coding Standards

## Design Principles (SOLID, KISS, DRY)

- **SOLID**:
  - **Single Responsibility (SRP)**: Give each file, struct, service, and React component one reason to change (e.g., separate `company_handler.go`, `opportunity_handler.go`, `opportunity_dto.go`, `company_repo.go`, `opportunity_repo.go`).
  - **Interface Segregation & Dependency Inversion (ISP / DIP)**: Keep inbound and outbound port interfaces narrow and role-focused, defined by the consumer in `internal/usecase/port/{inbound,outbound}` and implemented in `internal/adapter/<tech>`.
- **KISS (Keep It Simple)**: Favor direct, readable domain methods, straightforward SQL queries, and minimal state. Build only what the current ticket/spec requires without speculative abstractions or unused helpers.
- **DRY (Don't Repeat Yourself)**: Keep each domain invariant, validation rule, struct field group, SQL row hydration, and UI form field group in a single authoritative place; embed or extract shared helpers instead of duplicating logic across create/update paths.

## Architecture & Module Boundaries (ADR-0003)

- **`internal/domain`**: Pure Go entities, read-model projections (`*View`), value objects, typed domain errors, and pure domain evaluators. Import only the Go standard library; keep struct fields free of `json:` or `db:` tags.
- **`internal/usecase`**: Application services/interactors orchestrating domain aggregates and outbound ports via `TxManager.WithinTransaction(ctx, fn)`.
- **`internal/usecase/port/inbound` & `internal/usecase/port/outbound`**: Separate inbound and outbound port packages, one file per port/aggregate (`company_usecase.go`, `opportunity_repository.go`, `clock.go`, etc.). Declare only `interface` types inside `port/` packages; place concrete implementations in `internal/adapter/<tech>/`.
- **`internal/adapter/<tech>`**: Split files by single responsibility (`company_handler.go`, `opportunity_handler.go`, `opportunity_dto.go`, `company_repo.go`, `opportunity_repo.go`, `tx.go`).
- **`internal/app` (Composition Root)**: Wire concrete adapters and use-case services inside `internal/app/container.go`. Keep `cmd/cockpit/main.go` limited to flag/env parsing, opening infrastructure, and starting `http.Server`.

## Testing

- **Table-driven Go tests**: Structure multi-case Go unit and integration tests as table-driven slices (`tests := []struct{ name string; ... }{ ... }`) iterated with `t.Run(tc.name, ...)`.
- **Pre-agreed test seams**:
  - `internal/domain`: Pure unit tests with zero I/O or mocks.
  - `internal/adapter/http`: Full `app.NewHTTPHandler` integration tests against a real migrated SQLite database in `t.TempDir()`, injecting deterministic fakes only at external outbound ports (`Clock`, `AssistantGateway`, `JobPageParser`).
  - `web/`: Workspace view tests (`@testing-library/react` + Vitest + `QueryClientProvider` + `MemoryRouter`) exercising the real API client over stubbed `/api/*` HTTP handlers.

## Timestamps & Ubiquitous Language

- **UTC on the backend, local time in the UI**: Normalize and store all backend timestamps in UTC (`now.UTC()`, `RFC3339` / `RFC3339Nano`). Format UTC ISO strings into the user's local timezone at render time in the frontend via `web/src/lib/date.ts`.
- **Glossary fidelity**: Name types, fields, variables, and UI labels using the exact canonical terms in `GLOSSARY.md`.
