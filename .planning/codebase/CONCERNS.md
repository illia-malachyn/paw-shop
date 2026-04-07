# Codebase Concerns

**Analysis Date:** 2026-04-07

## Error Handling Issues

**Unhandled JSON Decode Errors in Tests:**
- Issue: Multiple test files ignore errors when decoding JSON responses, making tests fragile and masking real failures
- Files: 
  - `internal/handler/bundles_test.go` (lines 68, 88, 136, 174)
  - `internal/handler/discounts_test.go` (lines 25, 49, 115, 149)
  - `internal/handler/products_test.go` (lines 44, 68)
- Impact: If JSON response is malformed, test silently continues with nil values, causing confusing assertion failures downstream
- Fix approach: Add `if err != nil { t.Fatalf(...) }` checks after every `json.NewDecoder().Decode()` call in tests

**Unhandled JSON Encode Errors in Handlers:**
- Issue: All handlers use `json.NewEncoder(w).Encode()` without checking for errors
- Files:
  - `internal/handler/products.go` (line 53)
  - `internal/handler/bundles.go` (lines 30, 67, 109)
  - `internal/handler/discounts.go` (lines 69, 91, 125)
- Impact: If encoding fails (e.g., out of memory, invalid types), the error is silently ignored and client receives incomplete response
- Fix approach: Check error returns and write appropriate 500 responses on encoding failure

**Unchecked Empty Response in Apply Command:**
- Issue: `ApplyDiscountCommand.Execute()` returns 0 when product ID not found, which is indistinguishable from a free product
- Files: `internal/discount/command.go` (line 26-27)
- Impact: Handler cannot distinguish between valid zero-price result and missing product
- Fix approach: Return error interface or use optional/maybe pattern; modify handler to check price existence before applying discount

## Concurrency & State Management Issues

**No Thread Safety in Handler State:**
- Issue: `DiscountHandler` maintains mutable state (`subject`, `history`, `observers` map) with no synchronization primitives
- Files: `internal/handler/discounts.go` (lines 13-18, 29-34)
- Impact: Concurrent HTTP requests can cause race conditions on `subject.prices`, `history.history`, and `observers` map mutations
- Fix approach: Add `sync.RWMutex` to `PriceSubject`, `CommandHistory`, and `DiscountHandler`; protect map reads/writes in `observers`

**No Thread Safety in PriceSubject:**
- Issue: `PriceSubject.prices` and `observers` maps are accessed without locks
- Files: `internal/notification/observer.go` (lines 14-15)
- Impact: Concurrent calls to `SetPrice()` and `Subscribe()` can cause map corruption
- Fix approach: Add RWMutex to `PriceSubject`; lock on all map accesses

**Global Handler Instance in Tests:**
- Issue: Tests reuse `NewDiscountHandler()` instances within single test but HTTP state persists across test runs
- Files: `internal/handler/discounts_test.go`
- Impact: Test order dependency - `TestHandleUndo()` relies on previous `TestHandleApply()` execution within same handler instance
- Fix approach: Each test should create fresh handler, or use table-driven tests with isolated state

## Route Handling Issues

**Route Overlap Risk:**
- Issue: `/api/products/` (line 23 in main.go) registers a prefix handler that will match `/api/products` GET requests
- Files: `cmd/server/main.go` (lines 17, 23)
- Impact: In Go's DefaultServeMux, handlers registered with `/api/products` (exact path) and `/api/products/` (prefix) can conflict; behavior depends on registration order
- Fix approach: Use explicit path parsing or switch to a router library (gorilla/mux, chi) that handles overlapping routes deterministically

## Data Validation Issues

**Insufficient Input Validation:**
- Issue: Bundle builder accepts any string for `dog_size`, `food_type`, `pack_size` without validating against allowed values
- Files: `internal/bundle/builder.go` (lines 26-44)
- Impact: Invalid size values (e.g., "xyz_size") pass validation and create nonsensical bundles
- Fix approach: Add enum-like validation; use constants for allowed sizes and validate in `SetDogSize()`, `SetFoodType()`, `SetPackSize()`

**No Validation on URL Path Parsing:**
- Issue: `HandleSubscribe()` splits URL path without bounds checking before accessing parts[3]
- Files: `internal/handler/discounts.go` (lines 104-109)
- Impact: Malformed URLs like `/api/products/subscribe` will cause incorrect productID extraction or array index panic if parts[3] doesn't exist
- Fix approach: Check `len(parts) >= 5` and validate productID is non-empty after extraction

**Missing Discount Value Validation:**
- Issue: Discount strategies accept any float64 without validating reasonableness
- Files: `internal/discount/strategy.go` (lines 13-47)
- Impact: Percent strategy accepts > 100%, fixed strategy accepts negative amounts creating negative prices
- Fix approach: Add validation in strategy constructors to reject invalid values; clamp results to [0, original_price]

## Memory and Data Flow Issues

**Builder State Not Reset on Error:**
- Issue: `BundleBuilder.Build()` resets internal bundle state only on success (line 62), not on validation failure
- Files: `internal/bundle/builder.go` (lines 47-64)
- Impact: If `Build()` is called multiple times with invalid data, the builder resets after first successful build, causing unexpected behavior
- Fix approach: Move reset to end of function or return error without resetting internal state

**Shallow Clone in Bundle Prototype:**
- Issue: `Bundle.Clone()` creates deep copy of Extras slice but Bundle itself is shallow copied
- Files: `internal/bundle/bundle.go` (lines 22-33)
- Impact: While Extras slice is safe, future changes to Bundle structure could expose shallow fields
- Fix approach: Explicit deep copy is good; document this in comments; consider making Clone() a factory function

**No Limits on Observer List:**
- Issue: `PriceSubject.Subscribe()` has no limit on observers per product
- Files: `internal/notification/observer.go` (line 44-45)
- Impact: Malicious clients could subscribe unlimited observers causing unbounded memory growth and O(n) notification time
- Fix approach: Add per-product observer limit (e.g., 1000); return error when exceeded

## Test Coverage Gaps

**Missing Handler Integration Tests:**
- Files: `internal/handler/` tests exist but lack end-to-end scenarios
- What's not tested:
  - Multiple sequential discounts on same product
  - Interaction between discounts and subscriptions (does subscription capture both prices?)
  - Bundle creation with all optional fields
- Risk: Integration bugs in handler-to-model data flow undetected
- Priority: High

**Missing Factory Edge Cases:**
- Files: No tests for `factory/brand_factory.go` or `factory/product_factory.go`
- What's not tested:
  - `GetBrandFactory("unknown")` returns nil - not tested
  - Product ID uniqueness across brands not verified
  - Null/empty string product fields
- Risk: Silent failures when factories return nil products
- Priority: Medium

**No Negative Test Cases for Builder:**
- Files: `internal/bundle/` has no tests
- What's not tested:
  - Invalid dog_size values accepted without error
  - Build() without calling SetDogSize or SetFoodType
  - Building same bundle twice from same builder
- Risk: Silent acceptance of invalid bundles
- Priority: Medium

**Missing Discount Strategy Edge Cases:**
- Files: `internal/discount/strategy_test.go` (60 lines) has basic tests
- What's not tested:
  - Percent > 100 results in negative price
  - Fixed discount > price results in 0 (tested), but behavior undocumented
  - BuyNGetOne with N=0 causes division issues
  - Very large floats causing precision loss
- Risk: Unexpected discount calculations at scale
- Priority: Medium

## Security Considerations

**No Input Size Limits:**
- Risk: Email validation accepts arbitrarily long strings in subscribe endpoint
- Files: `internal/handler/discounts.go` (line 114)
- Current mitigation: None
- Recommendations: Add max length check (e.g., 255 chars) for email field

**No Rate Limiting:**
- Risk: Malicious clients can spam discount applications or subscriptions
- Files: All handlers accept unlimited requests
- Current mitigation: None
- Recommendations: Add rate limiting middleware; consider request throttling

**JSON Injection Risk:**
- Risk: `map[string]interface{}` responses allow arbitrary JSON structure if handlers are extended
- Files: All response encoding lines
- Current mitigation: Limited scope currently
- Recommendations: Use typed response structs instead of map[string]interface{}; validate all values before encoding

## Scaling Limits

**In-Memory Observer Storage:**
- Current capacity: Unbounded
- Limit: Server memory; all price notifications lost on restart
- Files: `internal/handler/discounts.go` line 32, `internal/notification/observer.go` line 71
- Scaling path: Implement database backing; add observer persistence layer; consider event streaming (Redis pub/sub, Kafka)

**Single-Threaded Command History:**
- Current capacity: Unbounded history per handler instance
- Limit: Memory usage grows linearly with discount operations
- Files: `internal/discount/command.go` (line 41-51)
- Scaling path: Add max history limit; implement undo/redo stack with size cap; consider history persistence with cleanup

**Hardcoded Product Prices in Handler:**
- Current: Prices initialized once in `NewDiscountHandler()`
- Limit: 2 brands x 3 products hardcoded; new products require code change
- Files: `internal/handler/discounts.go` (lines 23-27)
- Scaling path: Load prices from configuration or database; implement product registry pattern

## Dependencies at Risk

**No External Dependencies (Bundled):**
- No npm/pip/cargo dependencies listed in `go.mod`
- Risk: Low - minimal dependency surface
- Impact: Higher maintenance burden for OOP patterns not in stdlib
- Migration plan: Only add dependencies if stdlib becomes limiting (e.g., router library)

## Architecture Fragility Issues

**Tight Coupling Between Handlers and Models:**
- Issue: `DiscountHandler` directly instantiates `PriceSubject` and `CommandHistory`; handlers know about internal notification observers
- Files: `internal/handler/discounts.go` (lines 21-33)
- Why fragile: Hard to test handlers independently; swapping implementations requires handler changes
- Safe modification: Use dependency injection; pass subjects/history as constructor parameters
- Test coverage: Handler tests create fresh instances each time but don't mock dependencies

**Observer Pattern Without Unsubscribe:**
- Issue: `PriceObserver` interface has no unsubscribe mechanism
- Files: `internal/notification/observer.go` (lines 7-10)
- Why fragile: Subscribers persist for handler lifetime; can't clean up watchers
- Safe modification: Add `Unsubscribe(productID, observer)` method to `PriceSubject`
- Test coverage: No tests verify observer cleanup

**Bundle Registry Shared Across Requests:**
- Issue: `BundleHandler` creates single `BundleRegistry` instance shared across all requests
- Files: `internal/handler/bundles.go` (lines 15-18)
- Why fragile: Template modifications persist across requests if Clone() fails to truly isolate
- Safe modification: Thread-safe registry with copy-on-read semantics
- Test coverage: Tests only verify read operations, not concurrent access

## Missing Critical Features

**No Product Inventory Management:**
- Problem: Products are hard-coded; no stock tracking, no unavailability handling
- Blocks: Real e-commerce scenarios; can't handle out-of-stock products
- File references: `internal/factory/brand_factory.go`, `internal/handler/products.go`

**No Discount History per Product:**
- Problem: Command history tracks only last operation; no per-product discount timeline
- Blocks: Cannot query "what was price on date X?"; no audit trail for price changes
- File references: `internal/discount/command.go`

**No Notification Delivery Confirmation:**
- Problem: Observers are called synchronously but no delivery status returned
- Blocks: Cannot confirm if email notifications actually sent; no retry mechanism
- File references: `internal/notification/observer.go`

**No Error Recovery in Handlers:**
- Problem: Handlers assume all operations succeed; no fallback or graceful degradation
- Blocks: Partial failures can leave system in inconsistent state
- File references: All handler files

---

*Concerns audit: 2026-04-07*
