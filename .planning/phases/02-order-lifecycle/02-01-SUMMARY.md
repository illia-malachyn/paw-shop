---
phase: 02-order-lifecycle
plan: 01
subsystem: order
tags: [state-pattern, iterator-pattern, order-lifecycle, go]
dependency_graph:
  requires: []
  provides: [order.OrderState, order.OrderCollection, order.OrderIterator]
  affects: [internal/order]
tech_stack:
  added: []
  patterns: [State, Iterator]
key_files:
  created:
    - internal/order/state.go
    - internal/order/iterator.go
    - internal/order/state_test.go
    - internal/order/iterator_test.go
  modified:
    - internal/order/order.go
decisions:
  - "State pattern: unexported `state` field on Order enforces transition rules; Status string kept in sync for JSON serialization and backward compat with MacroCommand"
  - "Iterator: index-based cursor (no channels/goroutines) per plan spec; filteredIterator advances index in HasNext() scan so Next() can be called immediately after"
  - "NewOrder constructor initializes state to &NewState{} and Status to 'new' as single source of truth"
metrics:
  duration: ~10 minutes
  completed_date: "2026-04-07"
  tasks_completed: 3
  files_changed: 5
requirements_satisfied:
  - LIFE-01
  - LIFE-02
  - LIFE-03
  - LIFE-05
  - LIFE-06
  - LIFE-11
  - LIFE-12
---

# Phase 2 Plan 01: Order State and Iterator Patterns Summary

**One-liner:** State pattern with 5 concrete state objects enforcing lifecycle transitions (New->Confirmed->Shipped->Delivered/Cancelled), plus index-cursor Iterator for full and filtered OrderCollection traversal.

## What Was Built

### Task 1: State Pattern (`internal/order/state.go`, `internal/order/order.go`)

- `OrderState` interface: `Name() string`, `Next(*Order) error`, `Cancel(*Order) error`
- Five concrete states: `NewState`, `ConfirmedState`, `ShippedState`, `DeliveredState`, `CancelledState`
- Each state encodes which transitions are legal and which return descriptive errors
- `Order` struct updated: added unexported `state OrderState` field with JSON tags; `Status` string kept in sync with `state.Name()` for serialization and backward compatibility with `MacroCommand`/`ReportGenerator`
- Added `NewOrder(id, items)` constructor, `Next() error`, `Cancel() error`, `GetState() OrderState` delegation methods
- `ConfirmOrderCommand`, `RejectOrderCommand`, `MacroCommand` left unchanged — they operate on `Order.Status` string directly, preserving Phase 1 batch endpoint compatibility

### Task 2: Iterator Pattern (`internal/order/iterator.go`)

- `OrderIterator` interface: `HasNext() bool`, `Next() *Order`
- `OrderCollection` with unexported `orders []*Order` slice; methods: `Add`, `CreateIterator`, `CreateFilteredIterator(status)`, `GetByID(id)`, `Count`
- `allIterator` (unexported): index-based cursor, traverses all orders
- `filteredIterator` (unexported): index-based cursor, `HasNext()` scans forward to find next matching status, then `Next()` returns it and advances

### Task 3: Unit Tests (`state_test.go`, `iterator_test.go`)

- `TestStateTransitions`: 10 table-driven subtests covering all 5 states, both `Next` and `Cancel`, verifying state name, `Status` field sync, error presence, and error message content
- `TestFullLifecycle`: sequential `Next()` walk from New to Delivered
- `TestCreateIterator`: full traversal and empty-collection edge case
- `TestCreateFilteredIterator`: single match, no match, multiple matches in insertion order
- `TestGetByID`: found and not-found cases
- `TestOrderCollectionCount`: empty and populated

All 26 tests pass. All Phase 1 tests (`TestConfirmOrderCommand`, `TestRejectOrderCommand`, `TestMacroCommand`, `TestGenerateReport`, etc.) still pass.

## Commits

| Task | Commit | Message |
|------|--------|---------|
| 1 | `0fffaef` | feat(02-01): implement State pattern — OrderState interface and 5 concrete states |
| 2 | `f549402` | feat(02-01): implement Iterator pattern — OrderIterator, OrderCollection, FilteredIterator |
| 3 | `9c11fa3` | test(02-01): add unit tests for State and Iterator patterns |

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None. All state transitions and iterator traversals are fully implemented and wired.

## Threat Flags

None. This plan introduces only domain logic with no HTTP endpoints or new trust boundaries. The `state` field is unexported, satisfying T-02-01 (Tampering). The iterator uses index-based cursors with no goroutines, satisfying T-02-02 (DoS).

## Self-Check

- [x] `internal/order/state.go` exists
- [x] `internal/order/order.go` contains `state OrderState`, `Next()`, `Cancel()`, `GetState()`, `NewOrder()`
- [x] `internal/order/iterator.go` exists with `OrderIterator`, `OrderCollection`, `allIterator`, `filteredIterator`
- [x] `internal/order/state_test.go` exists with `TestStateTransitions` and `TestFullLifecycle`
- [x] `internal/order/iterator_test.go` exists with iterator tests
- [x] `go test ./internal/order/... -count=1` exits 0 (26 tests pass)
- [x] All Phase 1 tests still pass

## Self-Check: PASSED
