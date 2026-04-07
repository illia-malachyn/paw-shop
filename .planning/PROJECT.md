# PawShop

## What This Is

PawShop is an educational online dog food store built in Go, designed to demonstrate OOP design patterns for a university course. The backend exposes a REST API with HTTP handlers, in-memory data, and no external dependencies (stdlib only). Each feature (GitHub issue) introduces specific design patterns applied organically to solve real problems.

## Core Value

Each feature must clearly demonstrate its assigned design patterns through working, tested code — patterns are the deliverable, not just the product functionality.

## Requirements

### Validated

- ✓ Product catalog with Factory Method and Abstract Factory — existing (#1)
- ✓ Custom bundles with Prototype and Builder — existing (#2)
- ✓ Discounts and notifications with Strategy, Observer, Command — existing (#3)
- ✓ Batch order actions with MacroCommand and Template Method — Phase 1 (#4)
- ✓ Order lifecycle with State, Iterator, Chain of Responsibility — Phase 2 (#5)

### Active
- [ ] Product search with Interpreter and support chat with Mediator (#6)
- [ ] Cart with undo (Memento) and order export (Visitor) (#7)
- [ ] Notifications facade, logging proxy, and output bridge (#8)

### Out of Scope

- Frontend SPA — static landing page is sufficient
- Database persistence — in-memory storage only
- Authentication/authorization — not needed for pattern demonstration
- External dependencies — stdlib only (Go 1.23)
- Deployment — local development only

## Context

- University semester 2, 2026 OOP course
- Repository: `illia-malachyn/paw-shop` on GitHub
- Each issue gets its own feature branch, PR with pattern descriptions, and merge to main
- Issues #1-3 already implemented and merged
- Existing packages: `internal/models`, `internal/factory`, `internal/bundle`, `internal/discount`, `internal/notification`, `internal/handler`
- HTTP server runs on `:8080` via `cmd/server/main.go`

## Constraints

- **Language**: Go 1.23, no external dependencies (stdlib only)
- **Git workflow**: One branch per issue (`feature/{N}-{name}`), one commit per issue, PR with pattern descriptions, merge via `gh pr merge --merge`
- **Testing**: Unit tests mandatory for business logic + HTTP handlers (via `httptest`)
- **Structure**: New packages in `internal/`, handlers in `internal/handler/`, routes registered in `cmd/server/main.go`
- **PR format**: Must include pattern descriptions (problem, solution, why here)

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Issues #4→#5 sequential | #4 creates `internal/order`, #5 extends it | — Pending |
| In-memory storage only | Educational project, no persistence needed | — Pending |
| One phase per issue | Each issue is self-contained with its own patterns | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-04-08 after Phase 2 completion*
