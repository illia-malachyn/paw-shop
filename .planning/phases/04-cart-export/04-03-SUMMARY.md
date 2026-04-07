---
phase: 04-cart-export
plan: "03"
subsystem: handler
tags: [memento, visitor, http-handler, cart, tdd]
dependency_graph:
  requires:
    - 04-01  # cart.Cart + cart.CartHistory (Memento)
    - 04-02  # export.ExportCart + JSONExportVisitor + TextReceiptVisitor (Visitor)
  provides:
    - CartHandler HTTP endpoints (add, remove, undo, get, export)
    - Route registration in cmd/server/main.go
  affects:
    - cmd/server/main.go
tech_stack:
  added: []
  patterns:
    - Memento: CartHistory.Push/Pop for pre-mutation snapshots + Restore on undo
    - Visitor: ExportCart called with JSONExportVisitor or TextReceiptVisitor per format param
key_files:
  created:
    - internal/handler/cart.go
    - internal/handler/cart_test.go
  modified:
    - cmd/server/main.go
decisions:
  - Rollback snapshot on failed remove: Push memento before RemoveItem, Pop+Restore if error to keep history consistent
  - Route multiplexer in single HandleCart method: matches existing OrderHandler.HandleOrders pattern
  - Return empty JSON array [] instead of null when cart.Items is nil: avoids null in API responses
metrics:
  duration: "~15 minutes"
  completed: "2026-04-07T22:22:11Z"
  tasks_completed: 2
  files_created: 2
  files_modified: 1
requirements:
  - CART-04
  - CART-05
  - CART-06
  - CART-10
  - CART-13
---

# Phase 04 Plan 03: Cart HTTP Handlers Summary

**One-liner:** CartHandler wires Memento undo (CartHistory) and Visitor export (JSONExportVisitor/TextReceiptVisitor) to five HTTP endpoints registered in main.go.

## What Was Built

`CartHandler` in `internal/handler/cart.go` exposes all five cart endpoints via a single `HandleCart` route multiplexer:

| Method | Path | Behaviour |
|--------|------|-----------|
| POST | /api/cart/add | Validates CartItem, saves Memento snapshot, adds item, returns updated cart |
| POST | /api/cart/remove | Saves Memento snapshot, removes item by product_id (404 if not found, snapshot rolled back) |
| POST | /api/cart/undo | Pops CartHistory, restores snapshot (400 if history empty) |
| GET | /api/cart | Returns current cart items as JSON array |
| GET | /api/cart/export?format=json\|text | Applies JSONExportVisitor or TextReceiptVisitor via ExportCart |

Routes `/api/cart` and `/api/cart/` registered in `cmd/server/main.go`.

## Task Commits

| Task | Commit | Description |
|------|--------|-------------|
| 1 | 04f869a | feat(04-03): implement CartHandler with all cart HTTP endpoints |
| 2 | 477b095 | test(04-03): add HTTP handler tests for all cart endpoints |

## Test Coverage

11 handler tests in `internal/handler/cart_test.go`:

- `TestHandleCartAdd` — 200, status=added, 1 item in cart
- `TestHandleCartAddInvalid` — empty product_id returns 400
- `TestHandleCartRemove` — 200, status=removed, cart empty
- `TestHandleCartRemoveNotFound` — nonexistent product_id returns 404
- `TestHandleCartGet` — returns JSON array with correct item count
- `TestHandleCartUndo` — restores pre-add state, cart empty
- `TestHandleCartUndoEmpty` — fresh handler returns 400 with error message
- `TestHandleCartExportJSON` — valid JSON body, application/json Content-Type
- `TestHandleCartExportText` — contains "Receipt" and "Total", text/plain Content-Type
- `TestHandleCartExportInvalidFormat` — format=xml returns 400
- `TestHandleCartWrongMethod` — GET on /api/cart/add returns 405

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing correctness] Rollback Memento snapshot on failed remove**

- **Found during:** Task 1 implementation
- **Issue:** Plan specified Push before RemoveItem but did not address what happens when RemoveItem returns an error (item not found). Without rollback, a spurious snapshot pollutes CartHistory, causing the next undo to restore to a state that was never a real user action.
- **Fix:** After `RemoveItem` returns an error, immediately `Pop` the snapshot that was just pushed and `Restore` it, keeping history consistent.
- **Files modified:** internal/handler/cart.go (handleRemove)
- **Commit:** 04f869a

## Known Stubs

None — all endpoints return live cart state from in-memory `cart.Cart`.

## Threat Flags

No new network surfaces or trust-boundary changes beyond those declared in plan threat model (T-04-03 through T-04-06). Input validation for T-04-04 (product_id non-empty, price > 0, quantity > 0) is implemented in `handleAdd`.

## Self-Check: PASSED

| Check | Result |
|-------|--------|
| internal/handler/cart.go exists | FOUND |
| internal/handler/cart_test.go exists | FOUND |
| 04-03-SUMMARY.md exists | FOUND |
| commit 04f869a exists | FOUND |
| commit 477b095 exists | FOUND |
