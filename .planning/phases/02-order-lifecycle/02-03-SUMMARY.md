---
phase: 02-order-lifecycle
plan: 03
subsystem: handler
tags: [state-pattern, iterator-pattern, chain-of-responsibility, http-handlers, go]
dependency_graph:
  requires: [02-01, 02-02]
  provides: [handler.OrderHandler.HandleStatus, handler.OrderHandler.HandleListOrders, handler.OrderHandler.HandleCreateOrder]
  affects: [internal/handler/orders.go, internal/handler/orders_test.go, cmd/server/main.go]
tech_stack:
  added: []
  patterns: [State (via order.Next/Cancel), Iterator (via OrderCollection), Chain of Responsibility (via NewValidationChain)]
key_files:
  created: []
  modified:
    - internal/handler/orders.go
    - internal/handler/orders_test.go
    - cmd/server/main.go
decisions:
  - "OrderHandler migrated from map[string]*order.Order to *order.OrderCollection; GetOrders() shim retained for backward compat with existing batch tests"
  - "HandleOrders acts as path/method router for /api/orders/* catch-all; /api/orders/batch registered before /api/orders/ in main.go to preserve priority"
  - "HandleStatus extracts order ID by stripping /api/orders/ prefix and /status suffix from URL path (stdlib only, no router)"
  - "HandleListOrders uses CreateFilteredIterator when ?status= query param is non-empty, CreateIterator otherwise"
  - "HandleCreateOrder uses fmt.Sprintf(order-%d, Count()+1) for ID generation; seeds 3 orders so next generated ID is order-4"
metrics:
  duration: ~15 minutes
  completed_date: "2026-04-07"
  tasks_completed: 2
  files_changed: 3
requirements_satisfied:
  - LIFE-04
  - LIFE-07
  - LIFE-10
  - LIFE-14
---

# Phase 2 Plan 03: HTTP Handler for Order Lifecycle Summary

**One-liner:** Three new HTTP handlers wire State, Iterator, and Chain of Responsibility patterns to PATCH/GET/POST order endpoints, with path-based routing in a stdlib-only catch-all handler.

## What Was Built

### Task 1: Updated OrderHandler + Route Registration (`internal/handler/orders.go`, `cmd/server/main.go`)

- `OrderHandler` struct changed: `orders map[string]*order.Order` replaced with `collection *order.OrderCollection` and `validator order.OrderValidator`
- `NewOrderHandler()` seeds 3 orders via `order.NewOrder()` + `collection.Add()` and builds the validation chain via `order.NewValidationChain()`
- `HandleBatch` adapted to use `collection.GetByID()` instead of map lookup; MacroCommand logic unchanged
- `HandleReport` adapted to use `collection.CreateIterator()` to gather orders for report generation
- `HandleStatus` — PATCH /api/orders/{id}/status: extracts ID from path, routes `action:"next"` to `order.Next()` and `action:"cancel"` to `order.Cancel()`; returns descriptive 400 on illegal transitions
- `HandleListOrders` — GET /api/orders: uses `CreateFilteredIterator(status)` or `CreateIterator()` based on `?status=` query param; returns JSON array
- `HandleCreateOrder` — POST /api/orders: validates through `h.validator.Validate()` (StockValidator -> AddressValidator -> PaymentValidator chain); returns 201 on success
- `HandleOrders` — path/method router for `/api/orders/` catch-all; delegates to the 3 new handlers
- `GetOrders()` shim retained: converts `OrderCollection` back to `map[string]*order.Order` so existing batch tests require no changes
- `cmd/server/main.go`: added `http.HandleFunc("/api/orders/", orderHandler.HandleOrders)` after the `/api/orders/batch` registration

### Task 2: HTTP Handler Tests (`internal/handler/orders_test.go`)

Added 13 new tests covering all new endpoints:

**HandleStatus (6 tests):**
- `TestHandleStatus_NextAdvancesToConfirmed` — PATCH with action:"next" returns 200, status="confirmed"
- `TestHandleStatus_NextTwiceAdvancesToShipped` — two consecutive next calls reach "shipped"
- `TestHandleStatus_CancelReturnsStatusCancelled` — action:"cancel" returns 200, status="cancelled"
- `TestHandleStatus_NotFoundReturns404` — nonexistent order ID returns 404
- `TestHandleStatus_InvalidActionReturns400` — unknown action returns 400
- `TestHandleStatus_WrongMethodReturns405` — GET on status path returns 405

**HandleListOrders (3 tests):**
- `TestHandleListOrders_ReturnsAll3Orders` — GET /api/orders returns array of 3
- `TestHandleListOrders_FilteredByStatusNew` — ?status=new returns 3 orders all with status "new"
- `TestHandleListOrders_FilteredByNonexistentStatusReturnsEmpty` — returns empty array

**HandleCreateOrder (4 tests):**
- `TestHandleCreateOrder_ValidDataReturns201` — valid body returns 201 with id, status, items
- `TestHandleCreateOrder_OutOfStockItemReturns400` — "out-of-stock-item" triggers stock validation error
- `TestHandleCreateOrder_EmptyAddressReturns400` — empty address triggers address validation error
- `TestHandleCreateOrder_ZeroAmountReturns400` — amount=0 triggers payment validation error

All 8 pre-existing batch and report tests preserved and passing.

## Commits

| Task | Commit | Message |
|------|--------|---------|
| 1 | `1da362e` | feat(02-03): implement HandleStatus, HandleListOrders, HandleCreateOrder + route registration |
| 2 | `d53938f` | test(02-03): add HTTP handler tests for order lifecycle endpoints |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Cherry-picked state.go, iterator.go, and tests from orphaned branch**
- **Found during:** Pre-execution branch check
- **Issue:** The merge commit `3ce841e` (the required base) merged only the CoR branch into main but did not include the State/Iterator implementation commits (`0fffaef`, `f549402`, `9c11fa3`) which were on an orphaned branch. `internal/order/state.go` and `internal/order/iterator.go` did not exist in the worktree after rebasing.
- **Fix:** Cherry-picked all three Plan 01 implementation commits onto the worktree branch before starting Plan 03 implementation.
- **Files modified:** `internal/order/state.go`, `internal/order/iterator.go`, `internal/order/state_test.go`, `internal/order/iterator_test.go`
- **Commits:** `6b68526`, `9954bd5`, `6b2c662`

**2. [Rule 2 - Missing critical functionality] GetOrders() compatibility shim**
- **Found during:** Task 1
- **Issue:** Existing batch tests call `h.GetOrders()` which returned `map[string]*order.Order`. Migrating to `OrderCollection` would break them.
- **Fix:** Kept `GetOrders()` method but reimplemented it to iterate the collection and build a map, preserving test compatibility without changing test code.
- **Files modified:** `internal/handler/orders.go`

## Known Stubs

None. All handlers are fully wired: state transitions delegate to real `OrderState` implementations, listing uses real `OrderCollection` iterators, and creation runs through the real validation chain.

## Threat Flags

None beyond the threat model documented in the plan. T-02-05 (action validation) and T-02-06 (validation chain) are both implemented as required.

## Self-Check

- [x] `internal/handler/orders.go` contains `HandleStatus`, `HandleListOrders`, `HandleCreateOrder`, `HandleOrders`
- [x] `OrderHandler` uses `*order.OrderCollection` (not `map`)
- [x] `HandleStatus` extracts ID from path, delegates to `order.Next()` or `order.Cancel()`
- [x] `HandleListOrders` uses `CreateIterator()` or `CreateFilteredIterator()` based on `?status=`
- [x] `HandleCreateOrder` runs through `h.validator.Validate()` before creating order
- [x] `cmd/server/main.go` registers `/api/orders/` route after `/api/orders/batch`
- [x] `go build ./...` succeeds
- [x] `go test ./... -count=1` all pass (handler: 41 tests, order: all pass)
- [x] `go vet ./...` clean

## Self-Check: PASSED
