---
phase: 04-cart-export
plan: "01"
subsystem: cart
tags: [memento, cart, undo, deep-copy, go]
dependency_graph:
  requires: []
  provides: [internal/cart/cart.go, CartItem, Cart, CartMemento, CartHistory]
  affects: [04-02, 04-03]
tech_stack:
  added: []
  patterns: [Memento (Originator/Memento/Caretaker)]
key_files:
  created:
    - internal/cart/cart.go
    - internal/cart/cart_test.go
  modified: []
decisions:
  - "Cart.Items is exported (not unexported) to support JSON serialization and Visitor pattern access in Plan 02"
  - "CartMemento.items is unexported (opaque snapshot) to prevent caretaker from modifying internal state"
  - "Save() and Restore() use make+copy deep copy idiom — slice assignment would share backing array and corrupt snapshots"
metrics:
  duration: "~15 minutes"
  completed_date: "2026-04-07"
  tasks_completed: 2
  files_created: 2
---

# Phase 04 Plan 01: Cart Memento Pattern Summary

Cart domain model with undo support via Memento pattern: Cart (originator) + CartMemento (opaque snapshot) + CartHistory (LIFO caretaker stack) with deep-copy Save/Restore and 8 passing unit tests.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Implement Cart, CartMemento, CartHistory | 8195dcf | internal/cart/cart.go |
| 2 | Unit tests for Memento pattern | f84e78c | internal/cart/cart_test.go |

## What Was Built

- `CartItem` struct with json tags (`product_id`, `name`, `price`, `quantity`)
- `Cart` with `AddItem` (increments quantity on duplicate ProductID), `RemoveItem`, `UpdateQuantity`, `Save`, `Restore`
- `CartMemento` with unexported `items` field (opaque per Memento contract — caretaker cannot access internals)
- `CartHistory` LIFO stack with `Push`/`Pop`
- Both `Save()` and `Restore()` use `make` + `copy` deep copy, with `// deep copy — not assignment` comment per PITFALLS.md requirement

## Test Coverage

All 8 test functions pass:
1. `TestCartAddItem` — add item; add same ProductID increments quantity
2. `TestCartRemoveItem` — remove existing; error on non-existent
3. `TestCartUpdateQuantity` — update value; zero removes; error on non-existent
4. `TestCartSaveRestore` — save with 2 items, add third, restore → back to 2
5. `TestMementoDeepCopy` — THE critical test: save, mutate, restore proves no shallow copy corruption
6. `TestCartHistoryPushPop` — push 3, pop LIFO order; empty stack returns false
7. `TestUndoAfterAdd` — save empty, add, undo → empty cart
8. `TestUndoAfterRemove` — save 2 items, remove 1, undo → 2 items back

## Decisions Made

1. `Cart.Items` is exported — required for JSON marshaling in HTTP handlers (Plan 03) and Visitor pattern traversal (Plan 02). The plan explicitly specifies `Items []CartItem`.
2. `CartMemento.items` is unexported — enforces Memento pattern opaqueness. Caretaker (CartHistory) stores and returns mementos without inspecting or modifying state.
3. Value receiver return for `Save() CartMemento` (not pointer) — per STACK.md recommendation: value type prevents mutation after capture.

## Deviations from Plan

None - plan executed exactly as written.

## Threat Flags

No new threat surface introduced. T-04-01 (Tampering — CartMemento.items) mitigated by unexported field and deep copy in Save/Restore, as specified in the plan's threat model.

## Self-Check: PASSED

- `internal/cart/cart.go` exists: FOUND
- `internal/cart/cart_test.go` exists: FOUND
- Commit f84e78c (test RED phase): FOUND
- Commit 8195dcf (feat GREEN phase): FOUND
- All 8 tests pass: CONFIRMED (`go test ./internal/cart/ -v -count=1`)
- `go vet ./internal/cart/`: CLEAN
