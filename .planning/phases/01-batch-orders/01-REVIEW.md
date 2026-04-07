---
phase: 01-batch-orders
reviewed: 2026-04-07T00:00:00Z
depth: standard
files_reviewed: 7
files_reviewed_list:
  - cmd/server/main.go
  - internal/handler/orders.go
  - internal/handler/orders_test.go
  - internal/order/order.go
  - internal/order/order_test.go
  - internal/order/report.go
  - internal/order/report_test.go
findings:
  critical: 0
  warning: 3
  info: 4
  total: 7
status: issues_found
---

# Phase 01: Code Review Report

**Reviewed:** 2026-04-07
**Depth:** standard
**Files Reviewed:** 7
**Status:** issues_found

## Summary

Reviewed the batch-orders feature: `MacroCommand`/`ConfirmOrderCommand`/`RejectOrderCommand` (Command pattern), `GenerateReport` with `DailyReportGenerator`/`WeeklyReportGenerator` (Template Method pattern), HTTP handlers for `/api/orders/batch` and `/api/reports/{type}`, and their unit tests.

The Command and Template Method implementations are correct and well-tested. The main correctness concern is the lack of rollback in `HandleBatch` — when `MacroCommand.Execute()` fails mid-batch, orders processed before the failing entry are already mutated in memory. Three additional quality warnings and four informational items are noted.

---

## Warnings

### WR-01: Partial mutation with no rollback on batch error

**File:** `internal/handler/orders.go:66-68`

**Issue:** `MacroCommand.Execute()` halts at the first failing command and returns an error, but all commands executed before that point have already mutated their `Order.Status` in memory. The handler then returns `400 Bad Request`, leaving the in-memory store in a partially-applied state. For example, if `order_ids` is `["order-1", "order-2", "order-3"]` and `order-2` cannot be confirmed, `order-1` is silently left as `"confirmed"` even though the API response signals failure.

**Fix:** Validate all orders before executing any command, or implement rollback (Undo) after a failure. The simplest safe approach is a pre-flight validation pass:

```go
// Pre-flight: validate all orders before executing any command
for _, id := range req.OrderIDs {
    o, ok := h.orders[id]
    if !ok {
        http.Error(w, fmt.Sprintf("order not found: %s", id), http.StatusBadRequest)
        return
    }
    if o.Status != "new" {
        http.Error(w, fmt.Sprintf("order %s cannot be %sed: status is %q", id, req.Action, o.Status), http.StatusBadRequest)
        return
    }
}
// Only build and execute commands after all validations pass
```

---

### WR-02: Route subtree conflict between `/api/products` and `/api/products/`

**File:** `cmd/server/main.go:18,24`

**Issue:** Go's `net/http` mux treats a pattern ending in `/` as a subtree pattern that matches all paths with that prefix. `/api/products/` (line 24, for `HandleSubscribe`) will therefore also match the bare path `/api/products` via an automatic 301 redirect from the mux — which means a `GET /api/products` request may be redirected to `/api/products/` and then handled by `HandleSubscribe` instead of `HandleProducts`, depending on client redirect behavior.

**Fix:** Register the subscribe route with the full path pattern to avoid the subtree overlap:

```go
http.HandleFunc("/api/products/", discountHandler.HandleSubscribe) // keep subtree for {id}/subscribe
// Ensure /api/products is registered AFTER so the exact match wins over subtree for that path
http.HandleFunc("/api/products", productHandler.HandleProducts)
```

Or, better, use explicit path matching inside `HandleSubscribe` to reject requests that are not of the form `/api/products/{id}/subscribe`, and ensure `HandleProducts` is registered as an exact match (no trailing slash).

---

### WR-03: Unchecked error from `json.NewEncoder(w).Encode()`

**File:** `internal/handler/orders.go:72`, `internal/handler/orders.go:107`

**Issue:** The return value of `json.NewEncoder(w).Encode(...)` is silently discarded on both response paths. If encoding fails (e.g., an unencodable value is added later), the response will be silently incomplete or truncated without any server-side logging. CLAUDE.md permits silent ignoring in tests, but production handler code should at minimum log the error.

**Fix:**

```go
if err := json.NewEncoder(w).Encode(map[string]interface{}{
    "processed": len(req.OrderIDs),
    "action":    req.Action,
}); err != nil {
    // Response headers already sent; log for observability
    log.Printf("failed to encode response: %v", err)
}
```

---

## Info

### IN-01: No validation that `order_ids` is non-empty

**File:** `internal/handler/orders.go:50`

**Issue:** If the request body is `{"order_ids": [], "action": "confirm"}` or omits `order_ids` entirely, the handler silently returns `200 OK` with `"processed": 0`. This may be unexpected for clients and could mask client-side bugs.

**Fix:** Add a guard after decoding the request:

```go
if len(req.OrderIDs) == 0 {
    http.Error(w, "order_ids must not be empty", http.StatusBadRequest)
    return
}
```

---

### IN-02: Order status values are untyped magic strings

**File:** `internal/order/order.go:23,24,36,37`

**Issue:** The status values `"new"`, `"confirmed"`, and `"rejected"` are bare string literals repeated across `order.go`, `orders.go`, and tests. A typo in any future callsite would create an invalid status silently with no compile-time protection.

**Fix:** Define typed constants in the `order` package:

```go
const (
    StatusNew       = "new"
    StatusConfirmed = "confirmed"
    StatusRejected  = "rejected"
)
```

Then use `order.StatusNew`, `order.StatusConfirmed`, etc. throughout.

---

### IN-03: `GetOrders()` exposes the live internal map

**File:** `internal/handler/orders.go:114-116`

**Issue:** `GetOrders()` returns `h.orders` directly. Any caller can mutate the map or the `*Order` pointers it contains, bypassing handler logic entirely. Currently used only in tests, but the method is exported and accessible to any package that imports `handler`.

**Fix:** Return a shallow copy of the map, or restrict access to test-only helpers using build tags:

```go
func (h *OrderHandler) GetOrders() map[string]*order.Order {
    copy := make(map[string]*order.Order, len(h.orders))
    for k, v := range h.orders {
        copy[k] = v
    }
    return copy
}
```

---

### IN-04: `HandleReport` does not sanitize extra path segments

**File:** `internal/handler/orders.go:86`

**Issue:** `strings.TrimPrefix(r.URL.Path, "/api/reports/")` will yield `"daily/extra"` for a request to `/api/reports/daily/extra`, which falls to the `default` case and returns `400 Bad Request`. This is safe but produces a confusing error message ("unknown report type") for paths that are structurally wrong rather than semantically unknown.

**Fix:** Add a check for path separators in the extracted segment:

```go
reportType := strings.TrimPrefix(r.URL.Path, "/api/reports/")
if strings.Contains(reportType, "/") || reportType == "" {
    http.Error(w, "invalid report path", http.StatusNotFound)
    return
}
```

---

_Reviewed: 2026-04-07_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
