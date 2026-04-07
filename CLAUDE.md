<!-- GSD:project-start source:PROJECT.md -->
## Project

**PawShop**

PawShop is an educational online dog food store built in Go, designed to demonstrate OOP design patterns for a university course. The backend exposes a REST API with HTTP handlers, in-memory data, and no external dependencies (stdlib only). Each feature (GitHub issue) introduces specific design patterns applied organically to solve real problems.

**Core Value:** Each feature must clearly demonstrate its assigned design patterns through working, tested code — patterns are the deliverable, not just the product functionality.

### Constraints

- **Language**: Go 1.23, no external dependencies (stdlib only)
- **Git workflow**: One branch per issue (`feature/{N}-{name}`), one commit per issue, PR with pattern descriptions, merge via `gh pr merge --merge`
- **Testing**: Unit tests mandatory for business logic + HTTP handlers (via `httptest`)
- **Structure**: New packages in `internal/`, handlers in `internal/handler/`, routes registered in `cmd/server/main.go`
- **PR format**: Must include pattern descriptions (problem, solution, why here)
<!-- GSD:project-end -->

<!-- GSD:stack-start source:codebase/STACK.md -->
## Technology Stack

## Languages
- Go 1.23 - Backend server and core business logic
- HTML/CSS - Static frontend served from `static/` directory
- JavaScript - Client-side interactions in HTML
## Runtime
- Go 1.23
- Go Modules (go.mod)
- Lockfile: Not detected (no go.sum file committed)
## Frameworks
- Go standard library `net/http` - HTTP server and request handling
- No external web frameworks (e.g., no Chi, Gin, Echo)
- Go standard library `testing` - Built-in testing framework
- `net/http/httptest` - HTTP test utilities
- No build tool configuration detected (standard `go build`/`go run`)
## Key Dependencies
- `net/http` - HTTP server, handlers, and request/response management
- `encoding/json` - JSON encoding/decoding for API responses
- `strings` - String manipulation (URL parsing)
- `bytes` - Byte buffer operations
- `math` - Mathematical operations (likely for discount calculations)
- `fmt` - Formatted output and logging
- `log` - Basic logging
- Project uses only Go standard library
- go.mod contains only module declaration and Go version
## Configuration
- Hardcoded port: `:8080` in `cmd/server/main.go`
- No environment variable configuration detected
- Static files served from `./static/` directory
- Standard Go toolchain
- No Makefile, build.sh, or custom build configuration
## Platform Requirements
- Go 1.23+ installed
- Text editor or IDE with Go support
- Go 1.23+ runtime OR compiled binary
- Port 8080 must be available
- `./static/` directory must be accessible relative to binary execution
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

## Naming Patterns
- Lowercase with underscores for multi-word names: `products_test.go`, `brand_factory.go`
- Functional grouping: handlers, models, factories, discount, bundle, notification, factory
- Test files: `{module}_test.go` (co-located with implementation)
- PascalCase for exported functions: `NewProductHandler()`, `HandleProducts()`, `GetPrice()`
- camelCase for unexported functions: `resolveStrategy()`, `almostEqual()`
- Descriptive method names tied to responsibility: `Execute()`, `Undo()`, `Apply()`, `OnPriceChanged()`
- Handler functions follow pattern: `Handle{Feature}()` where Feature is capitalized
- camelCase for local variables and fields: `productID`, `newPrice`, `oldPrice`, `dogSize`
- Single-letter variables used only for loops: `i`, `p`, `w`, `r`, `t`, `f`, `b`, `s`, `h`, `o`
- Receiver variables use single letters: `h` for handlers, `s` for strategies/subjects, `b` for bundles/builders, `o` for observers
- Plural forms for slices: `products`, `observers`, `brands`, `extras`
- PascalCase for all types: `ProductHandler`, `PriceSubject`, `BundleBuilder`, `PercentStrategy`, `DryFood`
- Interface names end in "er" or "er" pattern: `PriceObserver`, `BrandFactory`
- Struct names are descriptive nouns: `Bundle`, `NotificationRecord`, `CommandHistory`
- Untyped string constants for HTTP methods: `http.MethodGet`, `http.MethodPost`
- Magic numbers embedded inline with clear context (e.g., `1 - s.Percent/100` in strategy)
- Status codes via `http.Status*` constants: `http.StatusOK`, `http.StatusBadRequest`, `http.StatusMethodNotAllowed`
## Code Style
- Go standard formatting (enforced by gofmt) - 1 tab indentation
- No explicit line length limit but typical ~100-110 characters observed
- Whitespace around operators: `a + b`, not `a+b`
- No spaces inside parentheses: `func()` not `func ( )`
- Go vet compliance
- No unused imports
- No unused variables or functions
- Exported functions require documentation
- Errors returned as last return value: `(*Bundle, error)`
- Early returns for error conditions
- `fmt.Errorf()` for creating formatted error messages
- Silent ignoring of errors with blank identifier where appropriate: `json.NewDecoder(w.Body).Decode(&products)` in tests
## Import Organization
- No aliases used; full paths always explicit
- Imports grouped naturally: stdlib first, then local
## Comments
- Comment exported types and functions (golint requirement)
- Comment package-level intent and patterns (design patterns like Builder, Observer, etc.)
- Inline comments explain "why" not "what": `// скасування знижки` explains intent, not action
- Ukrainian comments used throughout for domain context
- Exported functions: `// HandleProducts — GET /api/products`
- Type doc: `// ProductHandler — обробник HTTP-запитів для каталогу товарів.`
- Include HTTP method/path in handler docs
- Include design pattern name in structure docs
## Function Design
- Receivers use single letters (h, s, b, etc.)
- HTTP handlers follow signature: `(w http.ResponseWriter, r *http.Request)`
- Methods with data take receiver + 0-5 parameters
- Anonymous inline structs for request bodies in handlers:
- Single return for void-like operations (handlers modify `w` directly)
- `float64` for price calculations
- `(*T, error)` for builder pattern
- `bool` for existence checks: `ok` as variable name
- Multiple named returns not used; implicit ordering
## Module Design
- Capitalize first letter for all exported symbols: `NewProductHandler()`, `HandleProducts()`, `Product`
- Constructor functions follow `New{Type}()` pattern: `NewPriceSubject()`, `NewBundleBuilder()`
- Helper functions (unexported) use camelCase: `resolveStrategy()`, `almostEqual()`
- Each responsibility in own package: `handler`, `models`, `factory`, `discount`, `bundle`, `notification`
- Public interfaces in same package as implementations
- No cyclic imports (handlers depend on models, not vice versa)
- Handler types exported; internal state (fields) unexported but accessible via methods
- Strategy, Command, Observer interfaces are exported (used across packages)
- Factory interfaces are exported; concrete implementations exported
## API Response Patterns
- snake_case for JSON field names: `json:"product_id"`, `json:"dog_size"`, `json:"piece_count"`
- Lowercase struct fields for JSON marshaling: `ID`, `Name`, `Price`
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

## Pattern Overview
- Layered HTTP API with business logic separated from request handling
- Polymorphic product creation through abstract factories by brand
- Pluggable discount strategies with undo/redo command history
- Event-driven price change notifications via observer pattern
- Bundle creation through builder pattern and cloning through prototype pattern
## Layers
- Purpose: Accept HTTP requests, validate input, coordinate domain operations, return JSON responses
- Location: `internal/handler/`
- Contains: ProductHandler, BundleHandler, DiscountHandler
- Depends on: models, bundle, discount, notification, factory
- Used by: cmd/server/main.go
- Purpose: Core business operations independent of HTTP concerns
- Location: `internal/bundle/`, `internal/discount/`, `internal/notification/`
- Contains: Bundle, BundleBuilder, BundleRegistry, Strategy implementations, Command implementations, Observer implementations
- Depends on: models
- Used by: handler layer
- Purpose: Define product types and response structures
- Location: `internal/models/product.go`
- Contains: Product interface, DryFood, WetFood, Treat, ProductResponse
- Depends on: nothing
- Used by: handler, bundle, discount, factory
- Purpose: Abstract product creation by brand using factory pattern
- Location: `internal/factory/`
- Contains: BrandFactory interface, RoyalCaninFactory, AcanaFactory
- Depends on: models
- Used by: ProductHandler
- Location: `cmd/server/main.go`
- Triggers: Server startup
- Responsibilities: Initialize HTTP handlers, register routes, start listening on port 8080
## Data Flow
- **Product Prices:** Stored in DiscountHandler.subject (PriceSubject) as map[string]float64
- **Discount History:** Maintained in DiscountHandler.history (CommandHistory) as slice of Commands
- **Price Observers:** Registered in PriceSubject.observers as map[productID -> []PriceObserver]
- **Bundle Templates:** Stored in BundleRegistry.templates as map[string]*Bundle
- **User Notifications:** Cached in DiscountHandler.observers as map[email -> InMemoryObserver] for API access
## Key Abstractions
- Purpose: Polymorphic representation of any sellable item (dry food, wet food, treats)
- Examples: `internal/models/product.go` - DryFood, WetFood, Treat structs implement interface
- Pattern: Interface defines GetID(), GetName(), GetPrice(), GetCategory(), GetDetails() contract
- Purpose: Abstract family of products created by single brand
- Examples: `internal/factory/brand_factory.go` - RoyalCaninFactory, AcanaFactory
- Pattern: Each factory implementation creates one DryFood, one WetFood, one Treat
- Purpose: Pluggable discount calculation algorithms
- Examples: `internal/discount/strategy.go` - PercentStrategy, FixedStrategy, BuyNGetOneStrategy
- Pattern: Each strategy implements Apply(price float64) -> float64
- Purpose: Encapsulates discount operation with undo support
- Examples: `internal/discount/command.go` - ApplyDiscountCommand
- Pattern: Execute() applies discount, Undo() restores original price
- Purpose: React to price changes on specific products
- Examples: `internal/notification/observer.go` - LogObserver, InMemoryObserver
- Pattern: OnPriceChanged callback fires when PriceSubject.SetPrice() detects change
- Purpose: Step-by-step construction of Bundle with validation
- Location: `internal/bundle/builder.go`
- Pattern: Fluent interface (method chaining), returns *BundleBuilder for every setter except Build()
- Purpose: Central registry of template bundles available for cloning
- Location: `internal/bundle/templates.go`
- Pattern: Get() returns cloned prototype, List() returns available templates
## Error Handling
- **Invalid HTTP Method:** Early check `if r.Method != http.MethodPost` returns 405 MethodNotAllowed
- **Malformed JSON:** json.NewDecoder().Decode() returns error, handler responds with 400 BadRequest
- **Missing Required Fields:** Builder.Build() returns error if dog_size or food_type empty; BundleBuilder panics internally would not reach handler
- **Unknown Discount Type:** resolveStrategy() returns nil if type unrecognized; handler checks and returns 400
- **Empty Undo History:** CommandHistory.HasHistory() checked before Undo() call; returns 400 if false
- **Template Not Found:** BundleRegistry.Get() returns nil; handler checks and returns 404 NotFound
- **Price Floor:** FixedStrategy.Apply() ensures result never goes below 0
## Cross-Cutting Concerns
- Implementation: `internal/notification/observer.go` - LogObserver.OnPriceChanged() uses fmt.Printf
- Triggered by: PriceSubject.SetPrice() when oldPrice != newPrice
- HTTP input: Handler checks Method, parses JSON, validates required fields
- Bundle: BundleBuilder.Build() validates dog_size and food_type are non-empty
- Discount: resolveStrategy() validates discount_type is recognized
- Price: FixedStrategy floors result at 0 to prevent negative prices
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->
## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, or `.github/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->



<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
