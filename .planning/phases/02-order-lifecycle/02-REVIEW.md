---
phase: 02-order-lifecycle
reviewed: 2026-04-07T00:00:00Z
depth: standard
files_reviewed: 10
files_reviewed_list:
  - internal/order/state.go
  - internal/order/iterator.go
  - internal/order/validator.go
  - internal/order/order.go
  - internal/order/state_test.go
  - internal/order/iterator_test.go
  - internal/order/validator_test.go
  - internal/handler/orders.go
  - internal/handler/orders_test.go
  - cmd/server/main.go
findings:
  critical: 2
  warning: 2
  info: 2
  total: 6
status: issues_found
---

# Phase 02: Code Review Report

**Reviewed:** 2026-04-07
**Depth:** standard
**Files Reviewed:** 10
**Status:** issues_found

## Summary

The order lifecycle implementation introduces State, Iterator, Chain of Responsibility, and Command patterns. The state machine (`state.go`) and validation chain (`validator.go`) are well-structured and correct. The iterator (`iterator.go`) works correctly under standard usage patterns. The HTTP handler layer (`handler/orders.go`) is clean and consistent with project conventions.

Two critical bugs exist in `order.go`: both `ConfirmOrderCommand` and `RejectOrderCommand` bypass the State pattern entirely, directly mutating the `Status` string without updating the `state` field. This creates an irrecoverable desync between the state machine object and the exported status string — meaning orders that have been confirmed or rejected via `HandleBatch` can still be advanced or cancelled through `HandleStatus` as if they were still in the "new" state. There is also a non-atomic batch execution issue where a partial failure leaves the collection in a half-mutated state.

---

## Critical Issues

### CR-01: `ConfirmOrderCommand` bypasses the State machine, causing state/status desync

**File:** `internal/order/order.go:48-53`

**Issue:** `ConfirmOrderCommand.Execute()` sets `c.Order.Status = "confirmed"` directly but never calls `c.Order.Next()` or updates `c.Order.state`. After execution, `o.Status` is `"confirmed"` but `o.state` is still `&NewState{}`. A subsequent call to `HandleStatus` with action `"next"` will invoke `NewState.Next()` again, transitioning the order back to confirmed (overwriting state) and masking the double-advance. A call with action `"cancel"` will succeed (since `NewState.Cancel()` is valid), even though the order was already "confirmed" from the handler's perspective. The `HandleBatch` and `HandleStatus` endpoints are now inconsistent with each other for any order that passes through batch confirm.

**Fix:** Delegate to the state machine instead of mutating the field directly:
```go
func (c *ConfirmOrderCommand) Execute() error {
    if c.Order.Status != "new" {
        return fmt.Errorf("cannot confirm order %s: status is %q, expected \"new\"", c.Order.ID, c.Order.Status)
    }
    return c.Order.Next() // delegates to NewState.Next(), updates both state and Status
}
```

---

### CR-02: `RejectOrderCommand` introduces an orphaned "rejected" status with no corresponding state

**File:** `internal/order/order.go:61-67`

**Issue:** `RejectOrderCommand.Execute()` sets `c.Order.Status = "rejected"` but there is no `RejectedState` struct. After execution, `o.state` remains `&NewState{}` while `o.Status` is `"rejected"`. Any code that reads `o.GetState().Name()` (e.g., `filteredIterator`, `HandleStatus` response) returns `"new"`, not `"rejected"`, so `GET /api/orders?status=rejected` returns zero results even for rejected orders. More critically, `HandleStatus` can still advance or cancel a "rejected" order because the state machine has no idea the order was rejected — `NewState.Next()` and `NewState.Cancel()` both succeed.

**Fix:** Either add a `RejectedState` and route through it, or implement rejection as a cancel via the state machine. The simplest fix consistent with the existing state machine:
```go
// Option A: treat rejection as cancellation (simplest)
func (c *RejectOrderCommand) Execute() error {
    if c.Order.Status != "new" {
        return fmt.Errorf("cannot reject order %s: status is %q, expected \"new\"", c.Order.ID, c.Order.Status)
    }
    return c.Order.Cancel() // delegates to NewState.Cancel(), sets state=CancelledState, Status="cancelled"
}

// Option B: add RejectedState to state.go following the existing pattern,
// then call o.state = &RejectedState{}; o.Status = "rejected" consistently.
```

---

## Warnings

### WR-01: Batch execution is not atomic — partial failure leaves collection in a mutated state

**File:** `internal/handler/orders.go:69-73`

**Issue:** `MacroCommand.Execute()` stops at the first failure and returns the error, but any commands that already ran are not rolled back. If `order_ids` contains `["order-1", "order-2"]` and `order-1` succeeds but `order-2` fails (e.g., already confirmed), `order-1` is permanently mutated while the response sends HTTP 400. The caller has no way to know which orders were affected. For the current implementation with `ConfirmOrderCommand` and `RejectOrderCommand`, this will silently leave the in-memory store inconsistent.

**Fix:** Snapshot state before execution and restore on failure, or validate all commands before executing any:
```go
// Pre-validate: check all orders are in "new" state before executing
for _, id := range req.OrderIDs {
    o, ok := h.collection.GetByID(id)
    if !ok { /* already handled above */ }
    if o.Status != "new" {
        http.Error(w, fmt.Sprintf("order %s is not in 'new' state", id), http.StatusBadRequest)
        return
    }
}
// Only then build and execute the macro
```

---

### WR-02: `filteredIterator.HasNext()` mutates iterator state as a side effect

**File:** `internal/order/iterator.go:79-86`

**Issue:** `filteredIterator.HasNext()` advances `it.index` past non-matching entries. Calling `HasNext()` twice in a row without an intervening `Next()` call is safe only because the second call re-checks from wherever the first call stopped (which will still be a matching entry). However, calling `HasNext()` once, then calling `HasNext()` again after a different operation that does not call `Next()`, can silently skip entries. This also means that iterating without calling `HasNext()` first is undefined. The `allIterator` does not share this concern. While the standard `for it.HasNext() { it.Next() }` pattern works correctly, the contract violation (a predicate that mutates state) is a latent bug for any caller that calls `HasNext()` out of pattern.

**Fix:** Separate the "seek" concern from the "has next" predicate by caching the result of the seek:
```go
type filteredIterator struct {
    orders  []*Order
    status  string
    index   int
    peeked  bool // true when index already points to a valid match
}

func (it *filteredIterator) HasNext() bool {
    if it.peeked {
        return true
    }
    for it.index < len(it.orders) {
        if it.orders[it.index].GetState().Name() == it.status {
            it.peeked = true
            return true
        }
        it.index++
    }
    return false
}

func (it *filteredIterator) Next() *Order {
    if !it.HasNext() {
        return nil
    }
    o := it.orders[it.index]
    it.index++
    it.peeked = false
    return o
}
```

---

## Info

### IN-01: Order ID generation is fragile and can produce collisions

**File:** `internal/handler/orders.go:223`

**Issue:** `id := fmt.Sprintf("order-%d", h.collection.Count()+1)` generates IDs based on current collection size. The pre-seeded orders are "order-1", "order-2", "order-3", so the first `POST /api/orders` generates "order-4" (safe). However, there is no duplicate-check guard, so if the naming pattern were ever changed or orders were pre-populated differently, two orders could receive the same ID. `GetByID` returns the first match, silently hiding the duplicate.

**Fix:** Add a uniqueness check or use a counter that is independent of collection size:
```go
// Simple: check for collision before adding
for h.collection.GetByID(id) != nil {  // note: GetByID returns (*Order, bool)
    // increment counter
}
// Or use a dedicated auto-increment counter field on OrderHandler.
```

---

### IN-02: Unchecked `json.NewEncoder().Encode()` error in all handler responses

**File:** `internal/handler/orders.go:76, 113, 163, 192, 229`

**Issue:** All five JSON response writes use `json.NewEncoder(w).Encode(...)` without checking the returned error. The project convention (per CLAUDE.md) acknowledges silent ignoring of errors with the blank identifier in tests, but production handler code silently drops write errors. If the response write fails (e.g., client disconnected), no log entry is produced and the failure is invisible.

**Fix:** At minimum assign to `_` explicitly to signal intentional discard, or log the error:
```go
if err := json.NewEncoder(w).Encode(payload); err != nil {
    // response headers already sent; log for observability
    log.Printf("encode response: %v", err)
}
```
Given this is a university/educational project with no external dependencies, a simple `_ = json.NewEncoder(w).Encode(payload)` is also acceptable to signal explicit intent.

---

_Reviewed: 2026-04-07_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
