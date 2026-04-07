---
phase: 04-cart-export
plan: "02"
subsystem: export
tags: [visitor, json, text-receipt, cart-export]
dependency_graph:
  requires: [internal/cart/cart.go]
  provides: [internal/export/visitor.go, internal/export/visitor_test.go]
  affects: []
tech_stack:
  added: []
  patterns: [Visitor pattern with distinct method names per element type]
key_files:
  created:
    - internal/export/visitor.go
    - internal/export/visitor_test.go
  modified: []
decisions:
  - ExportCart convenience function chosen over requiring callers to drive Visitor manually
  - JSONExportVisitor.VisitCart is a no-op because total already accumulated in VisitItem
  - json.MarshalIndent used for readable JSON output
  - TextReceiptVisitor prepends header in Result() rather than VisitCart to keep separator/total in VisitCart
metrics:
  duration: ~10 minutes
  completed: "2026-04-07T22:18:41Z"
  tasks_completed: 2
  files_created: 2
  files_modified: 0
---

# Phase 04 Plan 02: Visitor Pattern Export Summary

**One-liner:** Visitor pattern with JSONExportVisitor and TextReceiptVisitor using distinct `VisitItem`/`VisitCart` method names (Go has no overloading), plus `ExportCart` convenience function and 7 unit tests.

## What Was Built

Two export visitors for `cart.Cart` implemented using the Visitor design pattern:

- **`OrderElement` interface** — `Accept(visitor OrderVisitor)` contract for elements
- **`OrderVisitor` interface** — three distinct methods: `VisitItem(item cart.CartItem)`, `VisitCart(c cart.Cart)`, `Result() string`
- **`CartItemElement` / `CartElement`** — wrappers enabling Visitor without modifying the cart package
- **`JSONExportVisitor`** — accumulates items with subtotals, produces indented JSON `{"items":[...],"total":N}`
- **`TextReceiptVisitor`** — accumulates lines, produces `=== Receipt ===` header, item lines `Name - Qty x Price = Subtotal`, separator, and `Total: N`
- **`ExportCart`** — convenience entry point; iterates items calling `VisitItem`, then calls `VisitCart`, returns `Result()`

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 (RED) | Failing unit tests for Visitor | 1a1ef33 | internal/export/visitor_test.go |
| 1 (GREEN) | Implement Visitor pattern | 8e7c6f6 | internal/export/visitor.go |

## Test Results

All 7 tests pass (`go test ./internal/export/ -v -count=1`):

- `TestJSONExportSingleItem` — JSON has correct subtotal and total
- `TestJSONExportMultipleItems` — JSON has all 3 items, correct total
- `TestJSONExportEmptyCart` — Valid JSON with empty items array and total 0
- `TestTextReceiptSingleItem` — Receipt header, item name, Total present
- `TestTextReceiptMultipleItems` — Both item names and correct total 350.00
- `TestTextReceiptEmptyCart` — Header and Total: 0.00
- `TestExportCartConvenience` — Both visitor types produce non-empty strings

`go vet ./internal/export/` — no warnings.

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None.

## Threat Flags

None. This plan implements domain logic only — no HTTP endpoints, no network boundaries, no PII. Threat model disposition was `accept` for JSON output (product names and prices are public catalog data).

## Self-Check: PASSED

- `internal/export/visitor.go` — FOUND
- `internal/export/visitor_test.go` — FOUND
- Commit `1a1ef33` — FOUND (RED: failing tests)
- Commit `8e7c6f6` — FOUND (GREEN: implementation)
