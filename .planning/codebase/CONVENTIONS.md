# Coding Conventions

**Analysis Date:** 2026-04-07

## Naming Patterns

**Files:**
- Lowercase with underscores for multi-word names: `products_test.go`, `brand_factory.go`
- Functional grouping: handlers, models, factories, discount, bundle, notification, factory
- Test files: `{module}_test.go` (co-located with implementation)

**Functions:**
- PascalCase for exported functions: `NewProductHandler()`, `HandleProducts()`, `GetPrice()`
- camelCase for unexported functions: `resolveStrategy()`, `almostEqual()`
- Descriptive method names tied to responsibility: `Execute()`, `Undo()`, `Apply()`, `OnPriceChanged()`
- Handler functions follow pattern: `Handle{Feature}()` where Feature is capitalized

**Variables:**
- camelCase for local variables and fields: `productID`, `newPrice`, `oldPrice`, `dogSize`
- Single-letter variables used only for loops: `i`, `p`, `w`, `r`, `t`, `f`, `b`, `s`, `h`, `o`
- Receiver variables use single letters: `h` for handlers, `s` for strategies/subjects, `b` for bundles/builders, `o` for observers
- Plural forms for slices: `products`, `observers`, `brands`, `extras`

**Types:**
- PascalCase for all types: `ProductHandler`, `PriceSubject`, `BundleBuilder`, `PercentStrategy`, `DryFood`
- Interface names end in "er" or "er" pattern: `PriceObserver`, `BrandFactory`
- Struct names are descriptive nouns: `Bundle`, `NotificationRecord`, `CommandHistory`

**Constants:**
- Untyped string constants for HTTP methods: `http.MethodGet`, `http.MethodPost`
- Magic numbers embedded inline with clear context (e.g., `1 - s.Percent/100` in strategy)
- Status codes via `http.Status*` constants: `http.StatusOK`, `http.StatusBadRequest`, `http.StatusMethodNotAllowed`

## Code Style

**Formatting:**
- Go standard formatting (enforced by gofmt) - 1 tab indentation
- No explicit line length limit but typical ~100-110 characters observed
- Whitespace around operators: `a + b`, not `a+b`
- No spaces inside parentheses: `func()` not `func ( )`

**Linting:**
- Go vet compliance
- No unused imports
- No unused variables or functions
- Exported functions require documentation

**Error Handling:**
- Errors returned as last return value: `(*Bundle, error)`
- Early returns for error conditions
- `fmt.Errorf()` for creating formatted error messages
- Silent ignoring of errors with blank identifier where appropriate: `json.NewDecoder(w.Body).Decode(&products)` in tests

## Import Organization

**Order:**
1. Standard library imports (`encoding/json`, `fmt`, `log`, `math`, `net/http`, `strings`, `testing`)
2. Local package imports from same module (`github.com/illia-malachyn/paw-shop/internal/...`)

**Path Aliases:**
- No aliases used; full paths always explicit
- Imports grouped naturally: stdlib first, then local

**Example from `internal/handler/discounts.go`:**
```go
import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/illia-malachyn/paw-shop/internal/discount"
	"github.com/illia-malachyn/paw-shop/internal/notification"
)
```

## Comments

**When to Comment:**
- Comment exported types and functions (golint requirement)
- Comment package-level intent and patterns (design patterns like Builder, Observer, etc.)
- Inline comments explain "why" not "what": `// скасування знижки` explains intent, not action
- Ukrainian comments used throughout for domain context

**Doc Comments:**
- Exported functions: `// HandleProducts — GET /api/products`
- Type doc: `// ProductHandler — обробник HTTP-запитів для каталогу товарів.`
- Include HTTP method/path in handler docs
- Include design pattern name in structure docs

**Pattern:**
```go
// TypeName — description of what this does.
type TypeName struct {}

// FunctionName — description of purpose.
// Additional details like HTTP method, path, or behavior.
func (r *Receiver) FunctionName() Type {
```

## Function Design

**Size:** Functions typically 20-110 lines; handlers run 35-75 lines including JSON marshaling

**Parameters:**
- Receivers use single letters (h, s, b, etc.)
- HTTP handlers follow signature: `(w http.ResponseWriter, r *http.Request)`
- Methods with data take receiver + 0-5 parameters
- Anonymous inline structs for request bodies in handlers:
  ```go
  var req struct {
      ProductID    string  `json:"product_id"`
      DiscountType string  `json:"discount_type"`
      Value        float64 `json:"value"`
  }
  ```

**Return Values:**
- Single return for void-like operations (handlers modify `w` directly)
- `float64` for price calculations
- `(*T, error)` for builder pattern
- `bool` for existence checks: `ok` as variable name
- Multiple named returns not used; implicit ordering

## Module Design

**Exports:**
- Capitalize first letter for all exported symbols: `NewProductHandler()`, `HandleProducts()`, `Product`
- Constructor functions follow `New{Type}()` pattern: `NewPriceSubject()`, `NewBundleBuilder()`
- Helper functions (unexported) use camelCase: `resolveStrategy()`, `almostEqual()`

**Package Structure:**
- Each responsibility in own package: `handler`, `models`, `factory`, `discount`, `bundle`, `notification`
- Public interfaces in same package as implementations
- No cyclic imports (handlers depend on models, not vice versa)

**Visibility:**
- Handler types exported; internal state (fields) unexported but accessible via methods
- Strategy, Command, Observer interfaces are exported (used across packages)
- Factory interfaces are exported; concrete implementations exported

## API Response Patterns

**Handler Response Structure:**
```go
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]interface{}{
    "field_name": value,
    "product_id": req.ProductID,
})
```

**JSON Tags:**
- snake_case for JSON field names: `json:"product_id"`, `json:"dog_size"`, `json:"piece_count"`
- Lowercase struct fields for JSON marshaling: `ID`, `Name`, `Price`

---

*Convention analysis: 2026-04-07*
