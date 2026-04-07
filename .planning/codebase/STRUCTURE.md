# Codebase Structure

**Analysis Date:** 2026-04-07

## Directory Layout

```
paw-shop/
├── cmd/                        # Command-line applications and entry points
│   └── server/                 # HTTP server application
│       └── main.go             # Server initialization and route setup
├── internal/                   # Private application code (not importable by external packages)
│   ├── models/                 # Data models and interfaces
│   │   └── product.go          # Product interface and implementations
│   ├── factory/                # Abstract factory implementations
│   │   ├── brand_factory.go    # BrandFactory interface and brand implementations
│   │   └── product_factory.go  # (if present, product creation helpers)
│   ├── bundle/                 # Bundle management (Builder, Prototype, Registry)
│   │   ├── bundle.go           # Bundle struct and Clone() method
│   │   ├── builder.go          # BundleBuilder for step-by-step construction
│   │   └── templates.go        # BundleRegistry for template management
│   ├── discount/               # Discount logic (Strategy, Command, History)
│   │   ├── strategy.go         # Strategy interface and implementations
│   │   ├── command.go          # Command interface and ApplyDiscountCommand
│   │   ├── strategy_test.go    # Tests for discount strategies
│   │   └── command_test.go     # Tests for discount commands
│   ├── notification/           # Price change notifications (Observer pattern)
│   │   ├── observer.go         # Observer interface and implementations
│   │   └── observer_test.go    # Tests for observers
│   └── handler/                # HTTP request handlers
│       ├── products.go         # GET /api/products handler
│       ├── bundles.go          # POST /api/bundles/* handlers
│       ├── discounts.go        # POST /api/discounts/* handlers
│       ├── products_test.go    # Tests for products handler
│       ├── bundles_test.go     # Tests for bundles handler
│       └── discounts_test.go   # Tests for discounts handler
├── static/                     # Static web assets
│   └── (HTML, CSS, JS files)   # Frontend files served by file server
├── go.mod                      # Go module definition
├── .gitignore                  # Git ignore rules
└── .git/                       # Git repository metadata
```

## Directory Purposes

**cmd/server/:**
- Purpose: Entry point for HTTP server application
- Contains: main.go with server initialization, route registration, and HTTP listener
- Key files: `cmd/server/main.go`

**internal/models/:**
- Purpose: Data model definitions and product interface contract
- Contains: Product interface, DryFood struct, WetFood struct, Treat struct, ProductResponse DTO
- Key files: `internal/models/product.go`

**internal/factory/:**
- Purpose: Abstract factory implementations for brand-specific product creation
- Contains: BrandFactory interface, RoyalCaninFactory, AcanaFactory, GetBrandFactory() lookup
- Key files: `internal/factory/brand_factory.go`

**internal/bundle/:**
- Purpose: Bundle creation, management, and cloning functionality
- Contains: Bundle struct, BundleBuilder (Builder pattern), BundleRegistry (Prototype storage)
- Key files: `internal/bundle/bundle.go`, `internal/bundle/builder.go`, `internal/bundle/templates.go`

**internal/discount/:**
- Purpose: Discount calculation and application with undo support
- Contains: Strategy interface, PercentStrategy, FixedStrategy, BuyNGetOneStrategy, Command interface, CommandHistory
- Key files: `internal/discount/strategy.go`, `internal/discount/command.go`

**internal/notification/:**
- Purpose: Price change notifications via observer pattern
- Contains: PriceObserver interface, PriceSubject (observable), LogObserver, InMemoryObserver, NotificationRecord DTO
- Key files: `internal/notification/observer.go`

**internal/handler/:**
- Purpose: HTTP request handlers coordinating request -> domain logic -> response flow
- Contains: ProductHandler, BundleHandler, DiscountHandler with their HTTP method handlers
- Key files: `internal/handler/products.go`, `internal/handler/bundles.go`, `internal/handler/discounts.go`

**static/:**
- Purpose: Serve static web assets (HTML, CSS, JavaScript)
- Contains: Frontend files served by http.FileServer at root path `/`
- Key files: None examined in this analysis

## Key File Locations

**Entry Points:**
- `cmd/server/main.go`: Server initialization, route setup, HTTP listener startup

**Product Catalog & Factory:**
- `internal/models/product.go`: Product interface definition and concrete types
- `internal/factory/brand_factory.go`: Brand factory implementations
- `internal/handler/products.go`: GET /api/products handler

**Bundle Management:**
- `internal/bundle/bundle.go`: Bundle struct with Clone() prototype implementation
- `internal/bundle/builder.go`: BundleBuilder with fluent interface
- `internal/bundle/templates.go`: BundleRegistry with default templates
- `internal/handler/bundles.go`: POST /api/bundles handlers

**Discount & Pricing:**
- `internal/discount/strategy.go`: Strategy implementations for different discount types
- `internal/discount/command.go`: Command pattern for discount application and undo
- `internal/handler/discounts.go`: POST /api/discounts/* handlers

**Notifications:**
- `internal/notification/observer.go`: Observer pattern implementation for price change notifications
- Integrated in: `internal/discount/command.go` (command notifies via PriceSubject)

**Testing:**
- `internal/handler/products_test.go`: Tests for ProductHandler
- `internal/handler/bundles_test.go`: Tests for BundleHandler
- `internal/handler/discounts_test.go`: Tests for DiscountHandler
- `internal/discount/strategy_test.go`: Tests for discount strategy implementations
- `internal/discount/command_test.go`: Tests for discount command pattern
- `internal/notification/observer_test.go`: Tests for observer pattern

## Naming Conventions

**Files:**
- Interface files: Named after the interface or concept (e.g., `strategy.go`, `observer.go`, `bundle.go`)
- Implementation files: Grouped with interface in same file (e.g., all strategies in `strategy.go`)
- Test files: Suffix with `_test.go` (e.g., `discounts_test.go` tests `discounts.go`)
- Handler files: Named after domain entity (e.g., `products.go`, `bundles.go`, `discounts.go`)
- Factory files: Named `*_factory.go` (e.g., `brand_factory.go`)

**Directories:**
- Lowercase with no hyphens or underscores (e.g., `internal`, `models`, `handler`)
- Plural for domain concepts (`internal/discount`, `internal/bundle`)
- `internal/` prefix follows Go convention for private packages

**Functions/Methods:**
- CamelCase starting with verb for handlers: `HandleProducts`, `HandleBuild`, `HandleApply`, `HandleUndo`, `HandleSubscribe`, `HandleClone`, `HandleTemplates`
- CamelCase starting with verb for constructors: `NewProductHandler`, `NewBundleBuilder`, `NewDiscountHandler`, `NewPriceSubject`, `NewBundleRegistry`, `NewCommandHistory`, `NewBundleBuilder`
- CamelCase for getters: `Get()`, `GetID()`, `GetName()`, `GetPrice()`, `GetDetails()`, `BrandName()`
- CamelCase for setters: `SetPrice()`, `SetName()`, `SetDogSize()`, `SetFoodType()`, `SetPackSize()`
- CamelCase for actions: `Apply()`, `Clone()`, `Build()`, `Subscribe()`, `Execute()`, `Undo()`, `List()`

**Types/Structs:**
- PascalCase: `Product`, `DryFood`, `WetFood`, `Treat`, `Bundle`, `BundleBuilder`, `PriceSubject`, `Strategy`, `Command`, `CommandHistory`, `BrandFactory`
- Interfaces end with "er" or are noun-based: `BrandFactory`, `Strategy`, `Command`, `PriceObserver`

**Constants/Enums:**
- Discount type strings: `"percent"`, `"fixed"`, `"buy_n_get_one"` (lowercase with underscores)
- Dog sizes: `"small"`, `"medium"`, `"large"`
- Food types: `"dry"`, `"wet"`, `"mixed"`
- Bundle pack sizes: `"standard"`, `"large"`, `"family"`

**Variables:**
- Lowercase with underscores for unexported struct fields: `oldPrice`, `newPrice`, `observers`, `templates`, `history`, `bundle`
- Uppercase for JSON tags: `json:"id"`, `json:"product_id"`, `json:"dog_size"`, `json:"extras"`

## Where to Add New Code

**New Feature:**
- Primary code: Place business logic in appropriate `internal/` subdirectory (e.g., new discount type -> `internal/discount/strategy.go`)
- Tests: Co-located with feature in same package, suffix with `_test.go`
- API endpoint: Add handler method to appropriate struct in `internal/handler/` and register route in `cmd/server/main.go`

**New Component/Module:**
- Implementation: Create new package under `internal/` (e.g., `internal/shipping/` for shipping feature)
- Organize by domain concept, not by layer (not `internal/models/shipping.go` and `internal/handlers/shipping.go`)
- Provide package-level interface and implementation struct
- Export only what external packages need

**Utilities:**
- Shared helpers: If used across multiple packages, create `internal/shared/` or appropriate domain package
- Package-local helpers: Keep as unexported functions in same file if only used within package
- Example: almostEqual() in `internal/discount/strategy_test.go` is unexported and only used in that test file

**Tests:**
- Location: Same package as code being tested (no separate `test/` directory)
- Naming: `*_test.go` in same directory as implementation
- Coverage: Aim for 100% coverage of handler methods; core logic methods (Apply, Execute, Clone, Build)
- Run with: `go test ./...` from project root

**Configuration:**
- Environment variables: Used in handlers if needed (not currently used; hardcoded product prices in DiscountHandler constructor)
- Add to: `cmd/server/main.go` if needed globally
- Pattern: Read from os.Getenv() at handler initialization time

## Special Directories

**internal/:**
- Purpose: Go convention marking these packages as private to this module (cannot be imported by external consumers)
- Generated: No
- Committed: Yes

**cmd/:**
- Purpose: Hold main applications (can have multiple commands/applications)
- Generated: No
- Committed: Yes
- Note: Each subdirectory (e.g., `cmd/server/`) has its own `main.go` and represents one executable

**static/:**
- Purpose: Static web assets served as-is without processing
- Generated: No (but may be updated during development)
- Committed: Yes
- Served by: `http.FileServer(http.Dir("./static"))` registered at "/" in main.go

**.git/:**
- Purpose: Git version control metadata
- Generated: Yes
- Committed: No (gitignore'd)

---

*Structure analysis: 2026-04-07*
