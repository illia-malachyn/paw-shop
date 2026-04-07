# Feature Landscape

**Domain:** Educational Go e-commerce — design pattern demonstration
**Researched:** 2026-04-07
**Issues in scope:** #4 (MacroCommand, Template Method), #5 (State, Iterator, CoR), #6 (Interpreter, Mediator), #7 (Memento, Visitor), #8 (Facade, Proxy, Bridge)

---

## What Makes an Implementation Educational vs a Toy

A toy implementation uses the pattern's names but doesn't expose its structural value. An educational implementation makes the *reason the pattern exists* visible through the domain problem it solves. The test: could you replace the pattern with a simpler approach and lose something real? If yes, the implementation is educational. If no, it's decoration.

---

## Issue #4 — MacroCommand + Template Method

### Table Stakes

| Feature | Why Required | Complexity | Notes |
|---------|-------------|------------|-------|
| `MacroCommand` implements the existing `Command` interface from `discount/command.go` | Pattern is only meaningful if MacroCommand is a Command — the composability is the point | Low | Must reuse `Execute() float64` and `Undo() float64` signatures |
| `MacroCommand.Execute()` runs all sub-commands in sequence and returns aggregate result | Without sequencing, it's just a slice wrapper | Low | Return type could be total value affected |
| `MacroCommand.Undo()` reverses all sub-commands in reverse order | Atomicity of rollback is the distinguishing behavior from simple loops | Low | Must undo last-first |
| At least two concrete commands composed in one macro (e.g., apply discount to product A + apply discount to product B) | One command in a MacroCommand proves nothing | Low | Use existing `ApplyDiscountCommand` as sub-commands |
| `BatchOrderReport` uses Template Method with a fixed skeleton in a base struct/embedded type | The invariant skeleton must be visible — if all steps live in the concrete type, Template Method isn't present | Medium | Go uses embedding, not classical inheritance |
| At least three concrete report types sharing the same skeleton (e.g., `SummaryReport`, `DetailedReport`, `CSVReport`) | Two types might look accidental; three makes the pattern intent clear | Medium | Each type overrides only the variant steps |
| Skeleton steps are broken out explicitly: `Header()`, `Body()`, `Footer()` or equivalent | If the template method calls only one overrideable function, it's just strategy by another name | Low | Step granularity demonstrates the pattern |
| HTTP endpoint `POST /api/orders/batch-discount` accepts a list of product IDs and discount specs | Batch operation needs a real trigger surface | Medium | Creates the `internal/order` package that #5 extends |

### Differentiators

| Feature | Value | Complexity |
|---------|-------|------------|
| `MacroCommand.Undo()` tracks which sub-commands actually succeeded and only undoes those | Demonstrates partial-failure awareness; toy implementations blindly undo all | Medium |
| Template Method steps have meaningful variance — e.g., `DetailedReport` Body iterates per-product, `SummaryReport` Body aggregates | If variants are just different strings, the pattern teaches nothing about step variance | Low |
| Report accepts an `io.Writer` parameter in the template method's `Generate` entry point | Idiomatic Go; makes reports testable without capturing stdout | Low |
| `CommandHistory` (from `discount`) accepts a `MacroCommand` — the composite is transparent to the history | Shows the Composite pattern property of MacroCommand without naming it | Low |

### Anti-Features

| Anti-Feature | Why to Avoid | What to Do Instead |
|-------------|-------------|-------------------|
| MacroCommand that only wraps commands of a single type (e.g., only `ApplyDiscountCommand`) | Loses the composability argument; looks like a batch discount, not a general macro | Accept `[]Command` interface slice |
| Template Method base with only one overrideable hook | Collapses to Strategy; doesn't show the invariant-skeleton vs variant-step distinction | Ensure at least 3 hook points in the skeleton |
| Creating a separate `Report` interface that each type implements directly | Defeats the purpose — Template Method exists precisely to avoid duplicating the skeleton in each type | Embed a base struct with the skeleton; override only the steps |
| Batch endpoint that calls individual discount apply endpoints internally | HTTP composition is not MacroCommand; the batch must happen at the Go type level | MacroCommand assembled and executed before any HTTP response |

### Dependencies

```
#4 creates internal/order (order model + batch handler)
#5 extends internal/order with State and Iterator
MacroCommand reuses discount.Command interface
Template Method report needs a product store to query
```

---

## Issue #5 — State + Iterator + Chain of Responsibility

### Table Stakes

| Feature | Why Required | Complexity | Notes |
|---------|-------------|------------|-------|
| Order has at least 4 states: `Pending`, `Confirmed`, `Shipped`, `Delivered`, plus `Cancelled` | Fewer states collapse the pattern to a boolean flag | Low | These are the natural e-commerce lifecycle states |
| Each state is a separate type implementing a `OrderState` interface with transition methods | State objects, not a switch in the order struct, is what defines the pattern | Medium | e.g., `Confirm()`, `Ship()`, `Deliver()`, `Cancel()` on the interface |
| Illegal transitions return an error, not a silent no-op | The pattern's safety guarantee is that the state controls what's possible; no error means no guard | Low | e.g., `Shipped.Confirm()` returns `ErrIllegalTransition` |
| State objects are set on the `Order` struct; `Order` delegates to its current state | The delegation is the structural heart of the pattern | Low | `order.Confirm()` calls `order.state.Confirm(order)` |
| `OrderIterator` implements `HasNext() bool` and `Next() *Order` over an order collection | Two methods is the minimal iterator contract | Low | Internal `[]Order` slice is not exported |
| HTTP endpoints trigger state transitions: `POST /api/orders/{id}/confirm`, `/ship`, `/deliver`, `/cancel` | Without an HTTP surface, the state machine is invisible to evaluators | Medium | Each returns current state in response |
| CoR has at least 3 handlers: `StockCheckHandler`, `PaymentValidationHandler`, `FraudDetectionHandler` | Fewer than 3 makes the chain look like an if-else | Medium | Each handler calls `next.Handle()` or stops chain |
| CoR handlers are linked at startup, not hardcoded in a function | The chain being configurable is the structural value | Low | Each handler holds a `next Handler` field |
| `POST /api/orders` passes a new order through the CoR before confirming | CoR must be in the critical path of order creation | Medium | If any handler rejects, order is not created |

### Differentiators

| Feature | Value | Complexity |
|---------|-------|------------|
| State transition methods accept the `Order` as a parameter so states can trigger side effects (e.g., send notification on `Shipped`) | Shows that State objects can act on their context, not just change it | Medium |
| `FilteringIterator` wraps the base iterator and skips orders by state | Iterator composition; demonstrates the pattern beyond simple traversal | Medium |
| CoR handler sets a reason string on rejection, returned in HTTP 400 body | Educational: makes which handler rejected visible for debugging/testing | Low |
| `GET /api/orders` uses the iterator internally (not `range` on a raw slice) | Makes iterator usage visible in production code path | Low |

### Anti-Features

| Anti-Feature | Why to Avoid | What to Do Instead |
|-------------|-------------|-------------------|
| State stored as a string enum on `Order` with a switch in each method | This is exactly the problem State solves; using it defeats the demonstration | State objects with delegation |
| Iterator that just returns the slice directly or wraps `range` trivially | Provides no encapsulation value | Ensure the internal data structure is not exported; iterator is the only access path |
| CoR where every handler always calls `next` (no handler ever short-circuits) | The whole point of CoR is conditional chain termination | At least one handler (e.g., stock check) must stop the chain on failure |
| CoR with handlers that have unrelated concerns combined in one struct | Loses the single-responsibility motivation for the pattern | One concern per handler |

### Dependencies

```
#5 extends internal/order created in #4
State transitions can fire Observer notifications (from notification package)
CoR stock check needs access to product catalog (from factory package)
Iterator is over the same order store used by batch in #4
```

---

## Issue #6 — Interpreter + Mediator

### Table Stakes

| Feature | Why Required | Complexity | Notes |
|---------|-------------|------------|-------|
| Search grammar supports at least: field filter (`category:dry`), price range (`price<500`), AND composition | One expression type is a lookup, not an interpreter | High | Defines the grammar hierarchy |
| `Expression` interface with `Interpret(context) bool` method | The interface is the contract that makes all expressions composable | Low | Context is a product |
| At least 3 expression types: `FieldExpression` (terminal), `RangeExpression` (terminal), `AndExpression` (non-terminal) | Terminal + composite mix is what makes it an Interpreter, not a filter function | Medium | `OrExpression` is a bonus |
| Parser function that takes a query string and returns an `Expression` tree | Without parsing, the expressions are just objects the caller assembles manually | High | e.g., `Parse("category:dry AND price<500")` |
| `GET /api/products/search?q=...` endpoint evaluates the expression tree against the product store | Real HTTP surface; response is filtered product list | Medium | |
| `Mediator` interface with a `Notify(sender, event, data)` method | The interface definition is what distinguishes mediator from direct method calls | Low | |
| At least 3 chat participants (`User`, `SupportAgent`, `BotAutoResponder`) that only hold a mediator reference, no reference to each other | The decoupling is the entire value proposition of the pattern | Medium | |
| `ChatRoom` struct implements `Mediator` and routes messages based on event type | The routing logic lives in the mediator, not the participants | Medium | |
| `POST /api/chat/send` and `GET /api/chat/messages` endpoints | HTTP surface makes the chat interaction testable by evaluators | Medium | In-memory message store is fine |

### Differentiators

| Feature | Value | Complexity |
|---------|-------|------------|
| `OrExpression` as a third composite type | Makes the grammar's recursive composition property visible | Low (once And exists) |
| Parser handles parentheses or operator precedence (AND before OR) | Shows the grammar is a real language, not just tokenization | High |
| `BotAutoResponder` participant replies automatically when a keyword is detected in mediator's `Notify` | Shows mediator can implement coordination logic, not just routing | Medium |
| Search endpoint returns which expression nodes matched (debug mode via query param) | Makes the expression tree evaluation visible — highly educational | Medium |

### Anti-Features

| Anti-Feature | Why to Avoid | What to Do Instead |
|-------------|-------------|-------------------|
| Interpreter with only one terminal expression type and no composite | This is a predicate function, not an interpreter | Add AND/OR composite expressions |
| Parser that only supports a single field filter (no boolean operators) | Loses the grammar hierarchy that defines the pattern | Support at least AND composition |
| Mediator that's just an Observer/EventBus re-skin with topic subscriptions | Mediator controls *how* participants interact; observer just notifies — the logic belongs in the mediator | Mediator routes and transforms messages, not just broadcasts |
| Chat participants that call mediator methods directly by name in response to other participants' messages | Shows coupling to the mediator's API, not decoupling between participants | Participants only emit events; mediator decides routing |
| Putting search filter logic in the HTTP handler instead of the Expression tree | Pattern is not demonstrated if it's bypassed | Handler only parses the query string and delegates entirely to the expression tree |

### Dependencies

```
Interpreter evaluates against Product (models.Product interface already exists)
Search endpoint needs access to product store (from factory/handler)
Mediator chat is self-contained in a new internal/chat package
No dependency on #4 or #5
```

---

## Issue #7 — Memento + Visitor

### Table Stakes

| Feature | Why Required | Complexity | Notes |
|---------|-------------|------------|-------|
| `Cart` struct (Originator) with `CreateMemento()` and `RestoreMemento(m)` methods | The originator/memento split is what defines the pattern | Low | |
| `CartMemento` struct that holds a deep copy of cart state (items + quantities) | Shallow copy breaks undo when items are modified after snapshot | Medium | Use value copies, not pointers |
| `CartHistory` (Caretaker) that stores a stack of mementos and exposes `Save()` and `Undo()` | The caretaker's ignorance of memento internals is a key structural requirement | Low | Caretaker must not inspect memento fields |
| Multiple undo steps (at least 3 add-item operations, each undoable independently) | Single-step undo can be faked without the pattern | Low | Tests must exercise 3 consecutive undos |
| `POST /api/cart/add`, `DELETE /api/cart/item`, `POST /api/cart/undo` endpoints | HTTP surface makes the undo feature real and testable | Medium | Cart is per-session; in-memory is fine |
| `CartItemVisitor` interface with `VisitDryFood`, `VisitWetFood`, `VisitTreat` methods | Type-specific visit methods (not a single `Visit(item)`) is what defines Visitor | Medium | One method per concrete cart item type |
| Each cart item type has an `Accept(CartItemVisitor)` method | Double-dispatch: the item's `Accept` calls the right `Visit*` method on the visitor | Low | This is the structural requirement |
| At least 2 concrete visitors: `PriceCalculatorVisitor` and one export format (e.g., `TextSummaryVisitor`) | One visitor proves nothing about extensibility — the point is adding behavior without modifying items | Medium | |
| `GET /api/cart/export?format=text` (or `json`) endpoint triggers visitor traversal | Real HTTP surface for visitor output | Medium | |

### Differentiators

| Feature | Value | Complexity |
|---------|-------|------------|
| `JSONExportVisitor` as a third visitor writing structured JSON | Shows that export format is a visitor concern; adding a format = adding a visitor | Medium |
| `CartMemento` stores a timestamp so history shows when each snapshot was taken | Makes the memento's role as a state record more concrete | Low |
| `CartHistory.Snapshot()` returns a list of memento summaries (count + total at each checkpoint) | Makes the caretaker's role visible in the API | Low |
| `PriceCalculatorVisitor` applies visitor-type-specific discount logic per item type | Shows visitor's ability to differentiate behavior by concrete type | Medium |

### Anti-Features

| Anti-Feature | Why to Avoid | What to Do Instead |
|-------------|-------------|-------------------|
| Memento that stores only the last state (no stack of history) | Single-step undo is achievable with one saved field; the stack is the pattern | Use a slice/stack of mementos |
| Visitor with a single `Visit(item interface{})` method that type-switches internally | This is not double-dispatch; it's a method with a type switch | Separate Visit methods per concrete type |
| Cart items that implement visitor logic themselves (checking visitor type) | Inverts the pattern; items should not know about visitors | Items only call `Accept`; visitors hold all type-specific logic |
| Combining Memento and Visitor in one `Cart` method | Obscures which pattern is doing what | Keep memento operations on `Cart`/`CartHistory`; visitor operations on item types and visitor structs |
| Exporting via a method on each item type rather than a visitor | Hardcodes export format in the domain model; adding a format requires modifying item types | All export logic belongs in visitor implementations |

### Dependencies

```
Cart items are Products from models package
Visitor requires concrete item types (DryFood, WetFood, Treat already exist in models)
No hard dependency on #4, #5, or #6
PriceCalculatorVisitor may read current prices from notification.PriceSubject
```

---

## Issue #8 — Facade + Proxy + Bridge

### Table Stakes

| Feature | Why Required | Complexity | Notes |
|---------|-------------|------------|-------|
| `NotificationFacade` wraps at least 3 subsystems: email sender, SMS sender, push notifier | Wrapping one subsystem is just a wrapper function, not a Facade | Medium | In-memory/stub implementations for each subsystem |
| Facade exposes a single `Notify(userID, event, message)` method to calling code | The simplified interface is the whole value; callers should not import subsystem packages | Low | Handler only imports the facade |
| Each subsystem has its own struct with its own interface — facade composes them internally | Subsystems must be real types, not ad-hoc function calls inside the facade | Low | Shows what the facade is hiding |
| `LoggingProxy` implements the same interface as the real service it wraps (e.g., `ProductService`) | Same interface is the structural requirement — callers can't tell they're talking to a proxy | Low | |
| Proxy adds logging before and after the real service call, transparently | Cross-cutting behavior without modifying the real service is the pattern's value | Low | Use the Bridge output (see below) for the logging output |
| `OutputBridge` separates `Logger` abstraction from `LogOutput` implementor | Two independent hierarchies — at least 2 Loggers (e.g., `InfoLogger`, `ErrorLogger`) and 2 Outputs (e.g., `StdoutOutput`, `BufferedOutput`) | Medium | The bridge is what allows combining them freely |
| `Logger` holds a `LogOutput` by interface, not by concrete type | If Logger directly instantiates its output, it's not a Bridge | Low | Set via constructor |
| `POST /api/orders/{id}/notify` endpoint exercises the facade | HTTP surface for the facade | Low | |
| Logging proxy wraps the product handler's underlying service; all product reads are logged | Makes the proxy's transparent interception visible | Medium | |

### Differentiators

| Feature | Value | Complexity |
|---------|-------|------------|
| `BufferedOutput` implementor that accumulates log lines, returned via `GET /api/logs` | Makes the bridge's output abstraction visible in the API; also makes tests trivial | Medium |
| Facade records which subsystems were notified per event in an audit log | Shows facade can add coordination logic, not just delegation | Low |
| Proxy wraps a cached service (read-through cache for products) rather than just logging | Demonstrates proxy as the general transparent-wrapper pattern, not just logging | High |
| `FileOutput` implementor that writes to a temp file (stdlib `os` only) | Shows bridge output hierarchy extending without touching Logger code | Low |

### Anti-Features

| Anti-Feature | Why to Avoid | What to Do Instead |
|-------------|-------------|-------------------|
| Facade that wraps only one subsystem | Indistinguishable from a simple wrapper function | Minimum 3 subsystems |
| Proxy that does not implement the same interface as the real subject | Not a proxy — it's a decorator with a different API | Proxy interface must match subject interface exactly |
| Bridge where both hierarchies have only one concrete type | Not a bridge — it's just an interface | Each hierarchy needs at least 2 concrete implementations |
| Using the existing `notification.PriceSubject` as the "facade" | PriceSubject is an Observer subject, not a facade over multiple notification channels | Build a new `NotificationFacade` that delegates to separate channel subsystems |
| Logger that directly references `os.Stdout` (hardcoded output) | Defeats the bridge's point | Logger must hold `LogOutput` by interface |

### Dependencies

```
NotificationFacade can notify on order events from #5
LoggingProxy wraps ProductService (built on top of factory package)
Bridge BufferedOutput feeds GET /api/logs endpoint
#8 is the most independent issue — no required dependency on #4-7
But LoggingProxy integrating with product reads makes it feel native
```

---

## Cross-Issue Feature Dependencies

```
#4 → #5: internal/order package created in #4, extended in #5
#5 → #8: order state transitions can trigger NotificationFacade
#7 → #6: search results feed into cart add (loose dependency, not structural)
#3 → #4: MacroCommand reuses discount.Command interface
#1 → #7: DryFood/WetFood/Treat concrete types required for Visitor's typed Accept methods
#1 → #6: Interpreter evaluates against models.Product
```

---

## MVP Recommendation per Issue

**Issue #4:** Prioritize: (1) MacroCommand composing existing Commands, (2) Template Method with 3 report types. Defer: partial-failure undo tracking — adds implementation complexity without much pattern clarity.

**Issue #5:** Prioritize: (1) State with 5 states + illegal transition errors, (2) CoR with 3 handlers on order create. Defer: FilteringIterator — base iterator is sufficient for pattern demonstration.

**Issue #6:** Prioritize: (1) Interpreter with AND + 3 expression types + parser, (2) Mediator ChatRoom with 3 participant types. Defer: operator precedence parsing — high complexity, low additional pattern insight.

**Issue #7:** Prioritize: (1) Memento stack with 3+ undo steps, (2) Visitor with typed Visit methods + 2 concrete visitors. Defer: JSONExportVisitor until text + price visitors are solid.

**Issue #8:** Prioritize: (1) Bridge with 2x2 hierarchy, (2) Facade over 3 subsystems. Defer: caching proxy — logging proxy demonstrates the pattern adequately.

---

## Sources

- Pattern analysis derived from: Gang of Four, "Design Patterns: Elements of Reusable Object-Oriented Software"
- Go-specific structural requirements derived from reading existing codebase (`internal/discount/command.go`, `internal/notification/observer.go`, `internal/models/product.go`)
- Educational vs toy distinction: analysis of what structural property each pattern provides that a simpler approach would not
