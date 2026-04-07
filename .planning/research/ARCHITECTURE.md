# Architecture Patterns

**Domain:** Go OOP design patterns e-commerce (educational)
**Researched:** 2026-04-07
**Confidence:** HIGH (based on direct codebase analysis)

---

## Existing Layer Structure (Issues #1–3)

The codebase uses four clean layers. Every new package must slot into this hierarchy — never skip layers, never introduce circular imports.

```
cmd/server/main.go              (Entry Point)
        |
        v
internal/handler/               (HTTP Handler Layer)
        |
        v
internal/{bundle,discount,notification,factory}/   (Domain Layer)
        |
        v
internal/models/                (Model/Data Layer — no deps)
```

**The rule proven by the existing code:** Domain packages import `models`. Handler packages import domain packages. `main.go` imports only `handler`. No domain package imports `handler`. `models` imports nothing.

---

## Recommended Architecture for New Packages

### How Each New Package Fits the Existing Layers

| New Package | Layer | Imports | Imported By |
|-------------|-------|---------|-------------|
| `internal/order` | Domain | `models`, `notification` | `handler`, (later) `cart`, `export` |
| `internal/cart` | Domain | `models`, `order` | `handler` |
| `internal/search` | Domain | `models`, `factory` | `handler` |
| `internal/chat` | Domain | `models` | `handler` |
| `internal/export` | Domain | `models`, `order`, `bundle`, `cart` | `handler` |
| `internal/notify` | Domain | `notification` | `handler`, optionally `order` |
| `internal/logging` | Cross-cutting | `models` (minimal) | injected into `handler`, `order`, `cart` via interfaces |

`logging` is the only package that breaks the strict domain pattern. It must be defined as an interface in the package that uses it (or in `models`) so that domain packages do not import `logging` directly — they accept a `Logger` interface. This prevents `logging` from creating import cycles when it inevitably needs to reference other packages.

---

## Component Boundaries

### `internal/order`

Introduced in issues #4 and #5. Issue #4 adds MacroCommand (batch order actions) and Template Method (order processing pipeline). Issue #5 extends it with State (order lifecycle), Iterator (iterate order items), and Chain of Responsibility (validation/processing chain).

**Responsibility:** Represent an order's lifecycle from creation to completion. Hold order items (product IDs + quantities), current state, and processing history.

**Key types to define:**
- `Order` struct with `State` field
- `OrderState` interface (`Name()`, `Next(*Order)`)
- Concrete states: `PendingState`, `ProcessingState`, `CompletedState`, `CancelledState`
- `OrderItem` struct: `{ProductID string, Quantity int, Price float64}`
- `OrderRepository` (in-memory map, created in `NewOrderHandler` — same pattern as `DiscountHandler.subject`)
- `OrderTemplate` (Template Method base for issue #4 batch processing)
- `MacroCommand` (wraps `[]Command` from existing `discount.Command` — or define a parallel `Command` interface here)
- `OrderIterator` for iterating items (issue #5)
- Chain links for order validation (issue #5)

**Does NOT import:** `handler`, `discount`, `bundle`

**Communicates with `notification`:** When an order transitions state, it can publish via `notification.PriceSubject` (reuse) or a new `OrderSubject`. Given the existing pattern, the cleanest approach is a new `OrderSubject` in `notification` that fires `OnOrderStateChanged` — this keeps notification as the single event bus.

### `internal/cart`

Introduced in issue #7. Patterns: Memento (undo cart changes).

**Responsibility:** Manage a shopping cart's contents. Track add/remove item history for undo.

**Key types:**
- `Cart` struct: `{Items []CartItem}`
- `CartItem` struct: `{ProductID string, Quantity int}`
- `CartMemento` struct: snapshot of `Items` (deep copy, same technique as `Bundle.Clone()`)
- `CartCaretaker`: holds `[]CartMemento` history stack, `Save()` and `Restore()` methods

**Imports:** `models` for product validation, `order` to convert a cart into an order

**Pattern mirror:** `CartCaretaker` is analogous to `CommandHistory`. The save/restore stack follows the same slice-as-stack idiom already proven in `discount.command.go`.

### `internal/search`

Introduced in issue #6. Pattern: Interpreter.

**Responsibility:** Parse and evaluate a query language for product filtering (e.g., `category:dry brand:acana price<1000`).

**Key types:**
- `Expression` interface: `Interpret(ctx SearchContext) bool`
- `SearchContext` struct: wraps a `models.Product` being evaluated
- Concrete expressions: `CategoryExpression`, `BrandExpression`, `PriceExpression`, `AndExpression`, `OrExpression`
- `QueryParser`: tokenizes a query string and builds an `Expression` tree
- `SearchEngine`: holds the product catalog slice, applies `Expression.Interpret()` across all products

**Imports:** `models` (to operate on `Product` interface), `factory` (to get the full product catalog, or accept `[]models.Product` injected by handler)

**Note on catalog access:** The product catalog currently lives only inside `handler/products.go` (instantiated per-request). For search to work, the catalog must be lifted into a shared store. The cleanest fix consistent with the existing architecture: create a `ProductCatalog` in `factory` or a thin `internal/catalog` package that both `handler/products.go` and `search` import. Do NOT have `search` import `handler`.

### `internal/chat`

Introduced in issue #6. Pattern: Mediator.

**Responsibility:** Coordinate messages between multiple participants (users, support bot) without direct coupling between them.

**Key types:**
- `ChatMediator` interface: `Send(from, message string)`
- `ChatRoom` struct: the concrete mediator, holds `map[string]Participant`
- `Participant` interface: `Receive(from, message string)`, `GetName() string`
- `UserParticipant` and `SupportBotParticipant` concrete implementations
- In-memory message log (slice of `Message` structs) for API retrieval

**Imports:** `models` (optional, if bot references products). No other domain package imports are needed — `chat` is self-contained.

**Does NOT import:** `order`, `cart`, `discount`

### `internal/export`

Introduced in issue #7. Pattern: Visitor.

**Responsibility:** Export order/cart/bundle data to different formats (JSON summary, plain text receipt, CSV) without modifying the visited structs.

**Key types:**
- `Visitable` interface: `Accept(Visitor)`
- `Visitor` interface: `VisitOrder(*order.Order)`, `VisitCart(*cart.Cart)`, `VisitBundle(*bundle.Bundle)`
- Concrete visitors: `JSONExporter`, `TextReceiptExporter`
- Each visitor accumulates output internally and exposes `Result() string`

**Imports:** `models`, `order`, `cart`, `bundle`

**This is the widest-importing package in the codebase.** It is acceptable because Visitor by definition must reference all visited types. The handler calls `export.NewJSONExporter()`, passes it to `order.Accept(exporter)`, and retrieves the result — the handler never touches the export format directly.

### `internal/notify`

Introduced in issue #8. Pattern: Facade.

**Responsibility:** Provide a single simplified interface to the notification subsystem, hiding `PriceSubject`, `LogObserver`, `InMemoryObserver` coordination that currently lives in `DiscountHandler`.

**Key types:**
- `NotificationFacade` struct: wraps `*notification.PriceSubject`
- `Subscribe(productID, email string)` — replaces the inline observer setup in `DiscountHandler.HandleSubscribe`
- `GetRecords(email string) []notification.NotificationRecord`
- `Notify(productID string, old, new float64)` — triggers subject directly (for testing)

**Imports:** `notification`

**Migration note:** `DiscountHandler` currently wires observers manually. After `notify` is introduced, `DiscountHandler` should delegate to `NotificationFacade`. This is a refactor within issue #8, not a breaking change — the HTTP API surface stays identical.

### `internal/logging`

Introduced in issue #8. Pattern: Proxy.

**Responsibility:** Wrap handler methods or domain operations with transparent logging, so callers see the same interface but all calls are recorded.

**Key design decision — avoid import cycle:**

Define the `Logger` interface in a place that both `logging` and domain packages can reference. Two options:

1. Define `Logger` in `internal/models` (already a zero-dependency package) — then any domain package can accept `models.Logger` without importing `logging`.
2. Define `Logger` in `internal/logging` itself, and have domain packages accept it as an injected interface parameter (constructor argument typed as an interface literal or a local interface redefinition).

Option 1 is cleaner given the existing `models` package is already the shared-contract location. Add `type Logger interface { Log(level, msg string) }` to `models/logger.go`.

**Key types in `logging`:**
- `StdoutLogger` struct: implements `models.Logger`, writes to `os.Stdout`
- `HandlerProxy` struct: wraps an `http.Handler`, logs method + path + duration before delegating
- `LoggingOrderService` (Proxy): wraps an order operation interface, logs each call

**Imports:** `models` (for `Logger` interface), `net/http` (stdlib only, for `HandlerProxy`)

---

## Data Flow: New Packages

### Issue #4 — Batch Order Actions

```
POST /api/orders/batch
  -> OrderHandler.HandleBatch()
  -> creates []order.OrderCommand (MacroCommand wraps them)
  -> MacroCommand.Execute() calls each sub-command
  -> Each sub-command: creates/updates an Order via OrderRepository
  -> Returns batch result JSON
```

### Issue #5 — Order Lifecycle

```
POST /api/orders
  -> OrderHandler.HandleCreate() -> Order{state: PendingState}
  -> OrderRepository.Save(order)

POST /api/orders/{id}/process
  -> OrderHandler.HandleProcess()
  -> chain.Handle(order) -> validation chain links execute
  -> order.state.Next(order) -> transitions to ProcessingState
  -> (optional) notification.OrderSubject.Notify(order)

GET /api/orders/{id}/items
  -> OrderHandler.HandleItems()
  -> order.Iterator() -> iterates []OrderItem
  -> Returns items as JSON array
```

### Issue #6 — Search + Chat

```
GET /api/search?q=category:dry+price<1000
  -> SearchHandler.HandleSearch()
  -> search.QueryParser.Parse(q) -> Expression tree
  -> search.SearchEngine.Filter(products, expression) -> []models.Product
  -> Returns filtered products JSON

POST /api/chat/send
  -> ChatHandler.HandleSend()
  -> chat.ChatRoom.Send(from, message)
  -> ChatRoom routes to participant.Receive()
  -> SupportBotParticipant may auto-reply via ChatRoom.Send()

GET /api/chat/messages
  -> ChatHandler.HandleMessages()
  -> chat.ChatRoom.GetLog() -> []Message
  -> Returns messages JSON
```

### Issue #7 — Cart + Export

```
POST /api/cart/add
  -> CartHandler.HandleAdd()
  -> caretaker.Save()  <- snapshot before change (Memento)
  -> cart.AddItem(productID, qty)

POST /api/cart/undo
  -> CartHandler.HandleUndo()
  -> caretaker.Restore() <- pops memento, restores cart

POST /api/cart/checkout
  -> CartHandler.HandleCheckout()
  -> order.NewOrderFromCart(cart) -> Order{state: PendingState}
  -> OrderRepository.Save(order)

GET /api/orders/{id}/export?format=json|text
  -> OrderHandler.HandleExport()
  -> export.NewJSONExporter() or export.NewTextExporter()
  -> order.Accept(exporter)
  -> Returns exporter.Result()
```

### Issue #8 — Notify Facade + Logging Proxy

```
POST /api/products/{id}/subscribe   (refactored, same URL)
  -> DiscountHandler.HandleSubscribe()
  -> notify.NotificationFacade.Subscribe(productID, email)
     (hides PriceSubject + observer wiring)

All HTTP requests (via proxy middleware):
  -> logging.HandlerProxy.ServeHTTP()
  -> records: method, path, timestamp
  -> delegates to wrapped handler
  -> records: duration, status code
```

The proxy is registered in `main.go` by wrapping the `http.ServeMux` or individual handlers:

```go
// main.go (issue #8 addition)
logger := logging.NewStdoutLogger()
http.Handle("/api/orders", logging.NewHandlerProxy(orderHandler, logger))
```

---

## Dependency Graph (Build Order)

Dependencies flow downward. Build bottom-up.

```
models                          <- build first, zero deps
    |
    +-- factory                 <- issue #1, done
    +-- notification            <- issue #3, done
    +-- bundle                  <- issue #2, done
    |
    +-- order                   <- issue #4/#5, depends on models + notification
    |       |
    |       +-- cart            <- issue #7, depends on models + order
    |       |
    |       +-- export          <- issue #7, depends on models + order + cart + bundle
    |
    +-- search                  <- issue #6, depends on models (+ factory indirectly)
    +-- chat                    <- issue #6, depends on models only
    |
    +-- notify                  <- issue #8, depends on notification
    +-- logging                 <- issue #8, depends on models (Logger interface)
    |
handler                         <- always last, imports all domain packages
    |
main.go                         <- imports only handler
```

**Strict build order for new packages:**

1. Extend `models` with `Logger` interface (if option 1 chosen for logging)
2. `internal/order` (no new-package deps; uses existing `notification`)
3. `internal/cart` (depends on `order`)
4. `internal/search` (parallel with `cart`, no dependency on it)
5. `internal/chat` (parallel with `search`, no dependencies on new packages)
6. `internal/export` (depends on `order`, `cart`, `bundle` — must come after all three)
7. `internal/notify` (depends only on `notification`, parallel with `export`)
8. `internal/logging` (depends only on `models`, parallel with `notify`)
9. Extend `internal/handler` with new handlers for all above
10. Register routes in `cmd/server/main.go`

Issues #4 and #5 must be sequential (PROJECT.md confirms this): #4 creates `internal/order` with the basic struct and MacroCommand/Template Method; #5 extends the same package with State/Iterator/CoR. No new package is created in #5.

---

## Patterns to Follow

### Constructor Injection (established by existing code)

All domain state lives in the handler struct, initialized in the `New*Handler()` constructor. Follow this exactly:

```go
type OrderHandler struct {
    repo     *order.OrderRepository
    notifier *notify.NotificationFacade
    logger   models.Logger
}

func NewOrderHandler(logger models.Logger) *OrderHandler {
    return &OrderHandler{
        repo:     order.NewOrderRepository(),
        notifier: notify.NewNotificationFacade(notification.NewPriceSubject()),
        logger:   logger,
    }
}
```

Do not use global variables or `init()` functions. The existing codebase has zero global mutable state outside handler constructors.

### In-Memory Repository Pattern

Every domain package that stores entities uses the same pattern established by `DiscountHandler`:

```go
type OrderRepository struct {
    orders map[string]*Order
}

func NewOrderRepository() *OrderRepository {
    return &OrderRepository{orders: make(map[string]*Order)}
}
```

Keys are string IDs. No persistence. No sync.Mutex needed (single-goroutine HTTP handler).

### Handler Method Naming

Follow the established `Handle*` prefix: `HandleCreate`, `HandleProcess`, `HandleItems`, `HandleBatch`, `HandleSearch`, `HandleSend`, `HandleMessages`, `HandleAdd`, `HandleUndo`, `HandleCheckout`, `HandleExport`, `HandleSubscribe` (already exists).

### HTTP Method Guard

Every handler starts with:

```go
if r.Method != http.MethodPost { // or Get
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
}
```

This is not optional — all existing handlers do it.

---

## Anti-Patterns to Avoid

### Anti-Pattern 1: Domain Package Importing Handler

`order` must not import `handler`. `cart` must not import `handler`. The existing codebase has zero such cycles — do not introduce the first one.

### Anti-Pattern 2: Shared Mutable Global State

The existing code stores all state inside handler structs (created in `main.go`). Adding `var globalOrderRepo = order.NewOrderRepository()` at package level would break the pattern and make tests unreliable. Keep state in constructor-initialized struct fields.

### Anti-Pattern 3: Putting Export Logic in the Visited Structs

Visitor pattern requires the export format logic to live in `export`, not in `order.Order.ToJSON()`. If you add format-specific methods to `Order`, you've defeated the pattern. `Order` implements `Accept(v Visitor)` only — format is the visitor's concern.

### Anti-Pattern 4: Chat Package Knowing About Orders

The Mediator pattern only works if `ChatRoom` is the only coordination point. `UserParticipant` and `SupportBotParticipant` must not import `order` or `cart`. If the support bot needs order info, the handler wires it at construction time via dependency injection, not by having `chat` import `order`.

### Anti-Pattern 5: Logging Package Creating Import Cycles

If `logging` imports `handler` to proxy handlers, and `handler` imports `logging` to use the proxy, you have a cycle. The solution: `logging.HandlerProxy` wraps `http.Handler` (stdlib interface), not a concrete handler type. Registration in `main.go` via wrapping is the only coupling point.

---

## Scalability Considerations (Educational Context)

All in-memory. No concurrency concerns. The patterns demonstrated are what matters.

| Concern | Current Approach | Pattern Constraint |
|---------|-----------------|-------------------|
| Order storage | map in handler struct | State pattern requires Order struct to own its state |
| Cart storage | map in handler struct (one cart per session concept) | Memento caretaker is per-cart |
| Search index | re-scan on each request | Interpreter builds tree per query — correct for pattern demo |
| Chat log | in-memory slice in ChatRoom | Mediator holds all state — correct |
| Export output | in-memory string in visitor | Visitor accumulates result — correct |

---

## Sources

- Direct analysis of `internal/discount/command.go`, `internal/notification/observer.go`, `internal/handler/discounts.go`, `internal/models/product.go`, `internal/bundle/bundle.go`, `cmd/server/main.go`
- PROJECT.md issue descriptions (#4–#8) for pattern assignments
- Go standard library documentation for `http.Handler` interface (used by logging proxy)
- Confidence: HIGH for all sections — based on direct codebase inspection, not training assumptions
