---
phase: 01-batch-orders
plan: "01"
subsystem: order
tags: [command-pattern, macro-command, template-method, tdd, go]
dependency_graph:
  requires: []
  provides: [internal/order]
  affects: [internal/handler]
tech_stack:
  added: [internal/order package]
  patterns: [MacroCommand, Template Method, Command]
key_files:
  created:
    - internal/order/order.go
    - internal/order/order_test.go
    - internal/order/report.go
    - internal/order/report_test.go
  modified: []
decisions:
  - "OrderCommand interface uses error return (not float64) — order commands change status, not prices; distinct from discount.Command interface"
  - "MacroCommand satisfies OrderCommand interface making it composable (composite pattern)"
  - "GenerateReport is a standalone function (not a method) so any type satisfying ReportGenerator can be used without embedding"
metrics:
  duration: ~10 minutes
  completed: 2026-04-07
  tasks_completed: 2
  files_created: 4
---

# Phase 1 Plan 01: Order Domain with MacroCommand and Template Method Summary

## One-liner

MacroCommand composing batch order confirm/reject commands and Template Method for daily vs weekly report generation, implemented TDD with 9 unit tests.

## What Was Built

Created the `internal/order` package establishing the order domain used by Issue #4 (and later extended by Issue #5). Two design patterns are demonstrated:

1. **MacroCommand pattern** (`order.go`): `ConfirmOrderCommand` and `RejectOrderCommand` each operate on a single `Order`. `MacroCommand` composes a slice of `OrderCommand` values and executes them sequentially — stopping and returning the first error encountered.

2. **Template Method pattern** (`report.go`): `GenerateReport` is the template function that calls `Header()`, `Body()`, and `Footer()` in a fixed order on any `ReportGenerator`. `DailyReportGenerator` produces an aggregate count summary; `WeeklyReportGenerator` produces per-order detail with status and items.

## Tasks

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Create Order model and MacroCommand pattern | cb40f60 | internal/order/order.go, internal/order/order_test.go |
| 2 | Create Template Method for report generation | abb85bb | internal/order/report.go, internal/order/report_test.go |

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| OrderCommand.Execute() returns error (not float64) | Order commands change state (status transitions), not numeric values; errors signal invalid transitions |
| MacroCommand itself satisfies OrderCommand | Enables composition — a MacroCommand can contain other MacroCommands |
| GenerateReport as standalone function (not receiver method) | Any value satisfying ReportGenerator can be passed without embedding; cleaner dependency injection |
| Status validation guards in ConfirmOrderCommand/RejectOrderCommand | Enforce valid state transitions — "new" only source state; prevents double-confirm/double-reject |

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None — all four files contain fully wired implementations with no placeholder data.

## Threat Flags

None — pure domain logic with no network endpoints, auth paths, file access, or external trust boundaries introduced.

## Self-Check: PASSED

Files verified:
- FOUND: internal/order/order.go
- FOUND: internal/order/order_test.go
- FOUND: internal/order/report.go
- FOUND: internal/order/report_test.go

Commits verified:
- FOUND: cb40f60 (feat(01-01): implement Order model and MacroCommand pattern)
- FOUND: abb85bb (feat(01-01): implement Template Method for report generation)

`go test ./internal/order/... -v` — 9 tests, all PASS
`go vet ./internal/order/...` — no warnings
