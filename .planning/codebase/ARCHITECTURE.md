# Architecture

**Analysis Date:** 2026-04-07

## Pattern Overview

**Overall:** Multi-pattern OOP architecture combining **Abstract Factory**, **Builder**, **Prototype**, **Command**, **Strategy**, and **Observer** design patterns

**Key Characteristics:**
- Layered HTTP API with business logic separated from request handling
- Polymorphic product creation through abstract factories by brand
- Pluggable discount strategies with undo/redo command history
- Event-driven price change notifications via observer pattern
- Bundle creation through builder pattern and cloning through prototype pattern

## Layers

**HTTP Handler Layer:**
- Purpose: Accept HTTP requests, validate input, coordinate domain operations, return JSON responses
- Location: `internal/handler/`
- Contains: ProductHandler, BundleHandler, DiscountHandler
- Depends on: models, bundle, discount, notification, factory
- Used by: cmd/server/main.go

**Domain/Business Logic Layer:**
- Purpose: Core business operations independent of HTTP concerns
- Location: `internal/bundle/`, `internal/discount/`, `internal/notification/`
- Contains: Bundle, BundleBuilder, BundleRegistry, Strategy implementations, Command implementations, Observer implementations
- Depends on: models
- Used by: handler layer

**Model/Data Layer:**
- Purpose: Define product types and response structures
- Location: `internal/models/product.go`
- Contains: Product interface, DryFood, WetFood, Treat, ProductResponse
- Depends on: nothing
- Used by: handler, bundle, discount, factory

**Factory/Creation Layer:**
- Purpose: Abstract product creation by brand using factory pattern
- Location: `internal/factory/`
- Contains: BrandFactory interface, RoyalCaninFactory, AcanaFactory
- Depends on: models
- Used by: ProductHandler

**Entry Point:**
- Location: `cmd/server/main.go`
- Triggers: Server startup
- Responsibilities: Initialize HTTP handlers, register routes, start listening on port 8080

## Data Flow

**Product Catalog Retrieval:**

1. Client requests GET `/api/products`
2. ProductHandler.HandleProducts() is invoked
3. Handler instantiates BrandFactory implementations (RoyalCaninFactory, AcanaFactory)
4. Each factory creates Product instances (DryFood, WetFood, Treat)
5. Handler transforms Product objects into ProductResponse DTOs via mapping loop
6. Response is JSON-encoded and returned to client

**Bundle Creation (Builder Pattern):**

1. Client POSTs to `/api/bundles` with JSON (name, dog_size, food_type, extras, pack_size)
2. BundleHandler.HandleBuild() receives request
3. NewBundleBuilder() creates builder instance with empty Bundle
4. Handler chains builder method calls: SetName() -> SetDogSize() -> SetFoodType() -> AddExtra() -> SetPackSize()
5. builder.Build() validates required fields (dog_size, food_type), sets defaults
6. Newly built Bundle is returned and JSON-encoded

**Bundle Cloning (Prototype Pattern):**

1. Client POSTs to `/api/bundles/clone` with template key and modifications
2. BundleHandler.HandleClone() retrieves template from BundleRegistry
3. Registry.Get() calls template.Clone() creating deep copy (Extras array is copied)
4. Handler modifies clone's Name and Extras if provided by client
5. Modified clone is returned separately from original template

**Discount Application (Strategy + Command + Observer):**

1. Client POSTs to `/api/discounts/apply` with product_id, discount_type, and value
2. DiscountHandler resolves Strategy implementation based on discount_type (PercentStrategy, FixedStrategy, BuyNGetOneStrategy)
3. Creates ApplyDiscountCommand wrapping the strategy and product reference
4. CommandHistory.Execute(cmd) calls cmd.Execute()
5. ApplyDiscountCommand.Execute():
   - Gets current price from PriceSubject
   - Saves old price for undo
   - Calls Strategy.Apply() to calculate new discounted price
   - Sets new price on PriceSubject
6. PriceSubject.SetPrice() detects price changed and notifies all subscribed PriceObservers
7. Observers (LogObserver, InMemoryObserver) execute OnPriceChanged callbacks
8. Response includes new_price and discount type

**Price Change Notifications (Observer Pattern):**

1. Client POSTs to `/api/products/{id}/subscribe` with email
2. DiscountHandler creates InMemoryObserver and LogObserver with user's email
3. Both observers are registered with PriceSubject for the product ID
4. When discount is applied, SetPrice() triggers notifications
5. Each observer independently records/logs the price change
6. API returns subscribed confirmation with product_id and email

**Undo Discount (Command History):**

1. Client POSTs to `/api/discounts/undo`
2. CommandHistory.Undo() pops last command from history stack
3. Calls cmd.Undo() which restores oldPrice on PriceSubject
4. PriceSubject detects change and notifies observers again
5. Returns restored price to client

**State Management:**

- **Product Prices:** Stored in DiscountHandler.subject (PriceSubject) as map[string]float64
- **Discount History:** Maintained in DiscountHandler.history (CommandHistory) as slice of Commands
- **Price Observers:** Registered in PriceSubject.observers as map[productID -> []PriceObserver]
- **Bundle Templates:** Stored in BundleRegistry.templates as map[string]*Bundle
- **User Notifications:** Cached in DiscountHandler.observers as map[email -> InMemoryObserver] for API access

## Key Abstractions

**Product Interface:**
- Purpose: Polymorphic representation of any sellable item (dry food, wet food, treats)
- Examples: `internal/models/product.go` - DryFood, WetFood, Treat structs implement interface
- Pattern: Interface defines GetID(), GetName(), GetPrice(), GetCategory(), GetDetails() contract

**BrandFactory Interface:**
- Purpose: Abstract family of products created by single brand
- Examples: `internal/factory/brand_factory.go` - RoyalCaninFactory, AcanaFactory
- Pattern: Each factory implementation creates one DryFood, one WetFood, one Treat

**Strategy Interface:**
- Purpose: Pluggable discount calculation algorithms
- Examples: `internal/discount/strategy.go` - PercentStrategy, FixedStrategy, BuyNGetOneStrategy
- Pattern: Each strategy implements Apply(price float64) -> float64

**Command Interface:**
- Purpose: Encapsulates discount operation with undo support
- Examples: `internal/discount/command.go` - ApplyDiscountCommand
- Pattern: Execute() applies discount, Undo() restores original price

**PriceObserver Interface:**
- Purpose: React to price changes on specific products
- Examples: `internal/notification/observer.go` - LogObserver, InMemoryObserver
- Pattern: OnPriceChanged callback fires when PriceSubject.SetPrice() detects change

**BundleBuilder:**
- Purpose: Step-by-step construction of Bundle with validation
- Location: `internal/bundle/builder.go`
- Pattern: Fluent interface (method chaining), returns *BundleBuilder for every setter except Build()

**BundleRegistry:**
- Purpose: Central registry of template bundles available for cloning
- Location: `internal/bundle/templates.go`
- Pattern: Get() returns cloned prototype, List() returns available templates

## Error Handling

**Strategy:** Defensive validation at handler entry points, graceful degradation in business logic

**Patterns:**

- **Invalid HTTP Method:** Early check `if r.Method != http.MethodPost` returns 405 MethodNotAllowed
- **Malformed JSON:** json.NewDecoder().Decode() returns error, handler responds with 400 BadRequest
- **Missing Required Fields:** Builder.Build() returns error if dog_size or food_type empty; BundleBuilder panics internally would not reach handler
- **Unknown Discount Type:** resolveStrategy() returns nil if type unrecognized; handler checks and returns 400
- **Empty Undo History:** CommandHistory.HasHistory() checked before Undo() call; returns 400 if false
- **Template Not Found:** BundleRegistry.Get() returns nil; handler checks and returns 404 NotFound
- **Price Floor:** FixedStrategy.Apply() ensures result never goes below 0

## Cross-Cutting Concerns

**Logging:** Via LogObserver in notification package - writes to stdout when price changes
- Implementation: `internal/notification/observer.go` - LogObserver.OnPriceChanged() uses fmt.Printf
- Triggered by: PriceSubject.SetPrice() when oldPrice != newPrice

**Validation:** Distributed across layers
- HTTP input: Handler checks Method, parses JSON, validates required fields
- Bundle: BundleBuilder.Build() validates dog_size and food_type are non-empty
- Discount: resolveStrategy() validates discount_type is recognized
- Price: FixedStrategy floors result at 0 to prevent negative prices

**Authentication:** Not implemented - API is open

**Authorization:** Not implemented - all operations available to all callers

---

*Architecture analysis: 2026-04-07*
