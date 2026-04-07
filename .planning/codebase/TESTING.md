# Testing Patterns

**Analysis Date:** 2026-04-07

## Test Framework

**Runner:**
- Go `testing` package (standard library)
- Run command: `go test ./...`
- Run specific test: `go test ./internal/handler -run TestHandleProducts_ReturnsAllProducts`
- Verbose mode: `go test -v ./...`
- Coverage: `go test -cover ./...`

**Assertion Library:**
- Standard library `testing.T` only
- Manual assertions with `t.Error()`, `t.Errorf()`, `t.Fatalf()`
- No external assertion framework (testify, etc.)

## Test File Organization

**Location:**
- Co-located with source files in same package
- Test files use `_test.go` suffix: `products_test.go`, `strategy_test.go`, `command_test.go`, `observer_test.go`

**Naming:**
- `Test{FunctionName}` for function tests: `TestHandleProducts_ReturnsAllProducts`
- `Test{TypeName}_{Behavior}` for struct/interface tests: `TestPercentStrategy`, `TestCommandHistory`
- Underscores separate test name from scenario: `TestHandleApply_Percent`, `TestHandleApply_UnknownType`
- Descriptive suffixes: `_ReturnsAllProducts`, `_MethodNotAllowed`, `_InvalidJSON`, `_EmptyHistory`

**File Structure Examples:**
- `internal/handler/products_test.go` - 111 lines, 5 test functions
- `internal/handler/discounts_test.go` - 171 lines, 8 test functions
- `internal/discount/strategy_test.go` - 60 lines, 5 test functions
- `internal/discount/command_test.go` - 96 lines, 3 test functions
- `internal/notification/observer_test.go` - 90 lines, 5 test functions

## Test Structure

**Suite Organization:**
Tests are NOT grouped in suites; each test function is independent and self-contained.

**Typical Handler Test Pattern:**
```go
func TestHandleProducts_ReturnsAllProducts(t *testing.T) {
	h := NewProductHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	w := httptest.NewRecorder()

	h.HandleProducts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var products []models.ProductResponse
	if err := json.NewDecoder(w.Body).Decode(&products); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(products) != 6 {
		t.Errorf("expected 6 products, got %d", len(products))
	}
}
```

**Key Characteristics:**
- No setup/teardown fixtures; instances created inline per test
- Handler tests use `httptest.NewRequest()` and `httptest.NewRecorder()`
- Tests decode JSON responses inline
- Single test responsibility (one or two assertions per test)
- Early exit with `Fatalf()` for critical failures
- `Errorf()` for non-fatal assertion failures

**Typical Strategy/Unit Test Pattern:**
```go
func TestPercentStrategy(t *testing.T) {
	s := &PercentStrategy{Percent: 10}
	if s.Name() != "percent" {
		t.Errorf("expected name 'percent', got '%s'", s.Name())
	}
	result := s.Apply(1000)
	if !almostEqual(result, 900) {
		t.Errorf("expected 900, got %.2f", result)
	}
}
```

**Typical Observer Test Pattern:**
```go
func TestPriceSubject_NotifiesObservers(t *testing.T) {
	s := NewPriceSubject()
	s.SetPrice("p1", 100)

	obs := &InMemoryObserver{UserEmail: "test@test.com"}
	s.Subscribe("p1", obs)

	s.SetPrice("p1", 80)

	if len(obs.Records) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(obs.Records))
	}

	r := obs.Records[0]
	if r.OldPrice != 100 || r.NewPrice != 80 {
		t.Errorf("expected 100->80, got %.2f->%.2f", r.OldPrice, r.NewPrice)
	}
}
```

## Fixtures and Factories

**Test Data:**
- No dedicated fixture files
- Test data created inline within test functions
- Handler tests instantiate fresh handler: `h := NewProductHandler()`
- Strategy tests create instances with parameters: `&PercentStrategy{Percent: 10}`
- Observer tests instantiate with email: `&InMemoryObserver{UserEmail: "test@test.com"}`

**Location:**
- Fixtures are co-located in `_test.go` files alongside tests
- No separate test data directory

**Helper Functions:**
- `almostEqual(a, b float64) bool` - used in strategy and command tests for float64 comparison
  ```go
  func almostEqual(a, b float64) bool {
      return math.Abs(a-b) < 0.01
  }
  ```

## Mocking

**What IS Mocked:**
- `http.ResponseWriter` via `httptest.NewRecorder()` - captures HTTP responses
- Request creation via `httptest.NewRequest()` - constructs fake HTTP requests

**What is NOT Mocked:**
- Domain objects (handlers, strategies, observers) - real implementations tested
- Interfaces - concrete implementations used in tests
- Subject/Observer relationships - full interaction tested with real implementations

**Pattern:**
```go
// Real handler instance
h := NewDiscountHandler()

// Fake HTTP request/response
req := httptest.NewRequest(http.MethodPost, "/api/discounts/apply", bytes.NewBufferString(body))
w := httptest.NewRecorder()

// Call handler (it modifies real response recorder)
h.HandleApply(w, req)

// Verify response
if w.Code != http.StatusOK {
    t.Fatalf("expected 200, got %d", w.Code)
}
```

## HTTP Handler Testing

**Request Construction:**
- `httptest.NewRequest(method, path, body)` for all handler tests
- Body as `nil` for GET requests
- Body as `bytes.NewBufferString(jsonString)` for POST with JSON payload

**Response Verification:**
```go
w := httptest.NewRecorder()
h.HandleApply(w, req)

// Check status
if w.Code != http.StatusOK {
    t.Fatalf("expected 200, got %d", w.Code)
}

// Check header
ct := w.Header().Get("Content-Type")
if ct != "application/json" {
    t.Errorf("expected application/json, got '%s'", ct)
}

// Decode and check body
var resp map[string]interface{}
json.NewDecoder(w.Body).Decode(&resp)
if resp["product_id"] != "rc-dry-01" {
    t.Errorf("expected product_id rc-dry-01, got %v", resp["product_id"])
}
```

## Test Categories

**Handler Tests (HTTP API):**
- Location: `internal/handler/`
- Test valid requests: method allowed, status 200, correct JSON response
- Test error cases: wrong method (405), invalid JSON (400), missing fields (400)
- Test data integrity: correct values in response, proper transformations applied

**Strategy Tests (Business Logic):**
- Location: `internal/discount/strategy_test.go`
- Test calculation correctness: `PercentStrategy`, `FixedStrategy`, `BuyNGetOneStrategy`
- Test edge cases: discount exceeds price (floor to 0), extreme percentages (50%)
- Test name() method for strategy identification

**Command Tests (Undo/Redo Pattern):**
- Location: `internal/discount/command_test.go`
- Test Execute() and Undo() methods
- Test state management: price saved, restored correctly
- Test Command History: sequential execute/undo, empty history edge case

**Observer Tests (Notification Pattern):**
- Location: `internal/notification/observer_test.go`
- Test price change detection
- Test observer notification (single and multiple observers)
- Test filter: only notify subscribed observers, only for subscribed products
- Test no-op: same price = no notification

## Assertion Patterns

**Status Code Assertions:**
```go
if w.Code != http.StatusOK {
    t.Fatalf("expected 200, got %d", w.Code)
}
```

**Float Comparison:**
```go
if !almostEqual(result, 900) {
    t.Errorf("expected 900, got %.2f", result)
}
```

**String Assertions:**
```go
if s.Name() != "percent" {
    t.Errorf("expected name 'percent', got '%s'", s.Name())
}
```

**Collection/Length Assertions:**
```go
if len(products) != 6 {
    t.Errorf("expected 6 products, got %d", len(products))
}
```

**Boolean Assertions:**
```go
if !brands["Royal Canin"] {
    t.Error("expected Royal Canin brand in catalog")
}
```

**Map/Key Existence:**
```go
price, ok := s.GetPrice("p1")
if !ok || price != 100 {
    t.Errorf("expected price 100, got %.2f (ok=%v)", price, ok)
}
```

**Decoded JSON Assertions:**
```go
var resp map[string]interface{}
json.NewDecoder(w.Body).Decode(&resp)
if resp["new_price"].(float64) != 1305 {
    t.Errorf("expected 1305, got %.2f", resp["new_price"].(float64))
}
```

## Error Testing

**Pattern:**
- Test that invalid HTTP method returns 405
- Test that invalid JSON returns 400
- Test that missing required fields returns 400
- Test that unknown discount types return 400
- Test that operations on empty state return 400

**Example:**
```go
func TestHandleApply_InvalidJSON(t *testing.T) {
	h := NewDiscountHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/discounts/apply", bytes.NewBufferString("{bad"))
	w := httptest.NewRecorder()

	h.HandleApply(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
```

## Coverage

**Requirements:** No explicit coverage requirements enforced

**Current Coverage:**
- Handler layer: 8 test functions covering main flows and error cases
- Discount strategies: 5 test functions covering all 3 strategy types + edge cases
- Command pattern: 3 test functions covering execute/undo/history
- Observer pattern: 5 test functions covering subscription, notification, filtering

**Total Test Functions:** 6 test files, ~26 test functions

**Gaps:**
- No tests for `ProductFactory` concrete implementations
- No tests for `BundleHandler.HandleBuild()` builder validation
- No tests for `BundleRegistry` template listing
- No negative tests for factory `GetBrandFactory()` with invalid brand

---

*Testing analysis: 2026-04-07*
