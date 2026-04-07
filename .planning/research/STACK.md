# Technology Stack: Design Patterns in Go 1.23

**Project:** PawShop — OOP design patterns educational e-commerce app
**Researched:** 2026-04-07
**Scope:** Implementing MacroCommand, Template Method, State, Iterator, Chain of
Responsibility, Interpreter, Mediator, Memento, Visitor, Facade, Proxy, Bridge
in Go 1.23 stdlib only

---

## Core Platform

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go | 1.23 | Language | Mandated by project; 1.23 adds first-class iter package, relevant for Iterator pattern |
| stdlib only | — | All dependencies | Mandated; sufficient for all 12 patterns |
| net/http | stdlib | HTTP routing | Already in use; no change needed |
| encoding/json | stdlib | JSON serialization | Already in use |

---

## Go Idioms That Map to Each Pattern

This is the core of what feeds the roadmap. Each section covers: the Go idiom to
reach for, what NOT to do (Java reflex), and confidence level.

---

### MacroCommand (Issue #4)

**Go idiom:** A `MacroCommand` struct holds a `[]Command` slice and iterates
through it in `Execute()`. Reverses the slice in `Undo()`. Implements the same
`Command` interface already defined in `internal/discount/command.go`.

```go
type MacroCommand struct {
    commands []Command
}

func (m *MacroCommand) Execute() float64 {
    var last float64
    for _, cmd := range m.commands {
        last = cmd.Execute()
    }
    return last
}

func (m *MacroCommand) Undo() float64 {
    var last float64
    for i := len(m.commands) - 1; i >= 0; i-- {
        last = m.commands[i].Undo()
    }
    return last
}
```

**Why this works:** The existing `Command` interface has `Execute() float64` and
`Undo() float64`. `MacroCommand` satisfies it, so `CommandHistory` can record
macros the same way it records atomic commands. No new interface needed.

**Do NOT:** Create a separate `MacroCommand` interface or use reflection to detect
composite vs atomic commands.

**Confidence:** HIGH — standard Composite + Command composition, verified against
existing codebase pattern.

---

### Template Method (Issue #4)

**Go idiom:** Go has no inheritance, so Template Method uses an interface of
"hook" methods plus a standalone function (or a method on a base struct) that
calls them in order. The standalone function IS the template; concrete types
satisfy the interface.

```go
// The interface defines the variable steps.
type OrderProcessor interface {
    Validate(order *Order) error
    Charge(order *Order) error
    Fulfill(order *Order) error
    Notify(order *Order) error
}

// The template is a plain function — not a method on a base struct.
func ProcessOrder(p OrderProcessor, order *Order) error {
    if err := p.Validate(order); err != nil {
        return err
    }
    if err := p.Charge(order); err != nil {
        return err
    }
    if err := p.Fulfill(order); err != nil {
        return err
    }
    return p.Notify(order)
}
```

**Why this works:** The calling code (handler or service) calls `ProcessOrder`
with any `OrderProcessor` implementation. Adding a new processor type means
implementing the interface, not overriding a base class method.

**Do NOT:** Embed a base struct and call `base.ProcessOrder()` from the embedded
type — Go embedding does not dispatch virtual method calls, so the base struct's
`ProcessOrder` would call base's own methods, not the embedding struct's overrides.
This is the single most common mistake when porting Template Method to Go.

**Confidence:** HIGH — confirmed against multiple sources including refactoring.guru
Go examples and community discussion on golang-nuts.

---

### State (Issue #5)

**Go idiom:** Define a `State` interface with all order lifecycle methods. The
`Order` struct holds a `currentState State` field and delegates to it. Each
concrete state struct implements the interface and changes `order.currentState`
when a transition occurs.

```go
type State interface {
    Next(o *Order) error
    Cancel(o *Order) error
    Name() string
}

type Order struct {
    ID           string
    currentState State
}

func (o *Order) Next() error   { return o.currentState.Next(o) }
func (o *Order) Cancel() error { return o.currentState.Cancel(o) }

type PendingState struct{}

func (s *PendingState) Next(o *Order) error {
    o.currentState = &ProcessingState{}
    return nil
}
func (s *PendingState) Cancel(o *Order) error {
    o.currentState = &CancelledState{}
    return nil
}
func (s *PendingState) Name() string { return "pending" }
```

**Why this works:** The State interface makes illegal transitions explicit —
calling `Next()` on a `DeliveredState` returns an error instead of panicking.
The `Order` struct never contains a `switch` on status strings.

**Do NOT:** Model state as a `const string` with a big `switch` inside every
order method. That is a procedural state machine, not the State pattern, and it
does not demonstrate the pattern clearly for OOP assessment.

**Confidence:** HIGH — standard GoF State in Go, verified against refactoring.guru
and John Doak's Go state machine article.

---

### Iterator (Issue #5)

**Go idiom:** Go 1.23 adds `iter.Seq[V]` and `iter.Seq2[K, V]` as first-class
iterator types that work with `for range`. Use these instead of the older
channel-based or `HasNext/Next` cursor approach.

```go
import "iter"

// OrderStore holds all orders; returns an iterator over them.
func (s *OrderStore) All() iter.Seq[*Order] {
    return func(yield func(*Order) bool) {
        for _, o := range s.orders {
            if !yield(o) {
                return
            }
        }
    }
}

// Caller:
for order := range store.All() {
    fmt.Println(order.ID)
}
```

For a filtered iterator (e.g., orders by status):

```go
func (s *OrderStore) ByStatus(status string) iter.Seq[*Order] {
    return func(yield func(*Order) bool) {
        for _, o := range s.orders {
            if o.currentState.Name() == status {
                if !yield(o) {
                    return
                }
            }
        }
    }
}
```

**Why iter.Seq over channel-based:** No goroutine scheduling overhead. Early
termination (break in the for-range loop) is handled by yield returning false
— no goroutine leak, no need for a done channel.

**Why NOT `HasNext/Next`:** That is a Java cursor idiom. Go 1.23 has a stdlib
type for this; using it is more idiomatic and more educational.

**Confidence:** HIGH — Go 1.23 official blog post "Range Over Function Types"
(go.dev/blog/range-functions) confirms this is the intended idiomatic pattern.

---

### Chain of Responsibility (Issue #5)

**Go idiom:** A `Handler` interface with `Handle(req *Request) error` and
`SetNext(Handler)`. Each concrete handler either handles the request or
forwards to the next handler in the chain.

```go
type OrderHandler interface {
    Handle(req *OrderRequest) (*Order, error)
    SetNext(h OrderHandler)
}

type BaseHandler struct {
    next OrderHandler
}

func (b *BaseHandler) SetNext(h OrderHandler) { b.next = h }

func (b *BaseHandler) PassToNext(req *OrderRequest) (*Order, error) {
    if b.next != nil {
        return b.next.Handle(req)
    }
    return nil, fmt.Errorf("no handler accepted request")
}
```

Concrete handlers embed `BaseHandler` to inherit `SetNext` and `PassToNext`,
only overriding `Handle`.

**For order validation context:** chain links could be: stock check →
fraud check → payment check → create order. Each link either returns an error
(rejecting) or calls `PassToNext`.

**Why NOT middleware-style func chains:** The `http.Handler` middleware style
(wrapping functions) works well for HTTP but loses the explicit "object with
state" quality needed to demonstrate the GoF pattern clearly for OOP assessment.

**Confidence:** HIGH — confirmed against refactoring.guru Go example and
multiple independent articles.

---

### Interpreter (Issue #6)

**Go idiom:** Define an `Expression` interface with `Interpret(ctx Context) bool`
(or a typed return). Build concrete expression nodes (terminal and non-terminal)
that form an AST. Parse a query string into this AST, then evaluate it.

```go
type Expression interface {
    Interpret(ctx SearchContext) bool
}

// Terminal: matches a single keyword
type KeywordExpression struct{ Keyword string }
func (k *KeywordExpression) Interpret(ctx SearchContext) bool {
    return strings.Contains(strings.ToLower(ctx.ProductName), strings.ToLower(k.Keyword))
}

// Non-terminal: AND
type AndExpression struct{ Left, Right Expression }
func (a *AndExpression) Interpret(ctx SearchContext) bool {
    return a.Left.Interpret(ctx) && a.Right.Interpret(ctx)
}

// Non-terminal: OR
type OrExpression struct{ Left, Right Expression }
func (o *OrExpression) Interpret(ctx SearchContext) bool {
    return o.Left.Interpret(ctx) || o.Right.Interpret(ctx)
}
```

The parser takes a query string like `"chicken AND large"` and builds the AST.
Evaluation iterates over products and applies the AST.

**Scope for this project:** Keep the grammar minimal — AND, OR, single keywords.
Do NOT parse full boolean algebra with grouping. That scope is correct for a
product search in a university project and keeps the parser implementable in
~100 lines of stdlib code.

**Confidence:** HIGH for pattern structure; MEDIUM for parser implementation
complexity estimate (keeping grammar simple is critical).

---

### Mediator (Issue #6)

**Go idiom:** A `Mediator` interface with a `Notify(sender Component, event string)`
method. Each `Component` holds a reference to the mediator and calls
`mediator.Notify(c, "event")` instead of calling other components directly.

```go
type Mediator interface {
    Notify(sender ChatComponent, event string, data string)
}

type ChatComponent interface {
    SetMediator(m Mediator)
}

type ChatRoom struct {
    users map[string]*ChatUser
}

func (r *ChatRoom) Notify(sender ChatComponent, event string, data string) {
    switch event {
    case "message":
        // broadcast to all other users
    case "join":
        // welcome the joining user
    }
}
```

**Why NOT channels:** Channels create implicit coupling through shared state
(the channel itself). The GoF Mediator's point is the mediator struct contains
all coordination logic visibly. For OOP assessment, a struct mediator is more
demonstrable than a pub/sub event bus.

**Confidence:** HIGH — verified pattern structure; MEDIUM on domain fit
(support chat is a plausible fit but simple enough to implement cleanly).

---

### Memento (Issue #7)

**Go idiom:** Three types: Originator (the cart), Memento (an opaque snapshot),
Caretaker (the undo stack). The Memento should be a value type (struct, not
pointer) to prevent mutation after capture.

```go
// Memento is intentionally opaque — unexported fields.
type CartMemento struct {
    items []CartItem  // deep copy
}

// Originator
type Cart struct {
    items []CartItem
}

func (c *Cart) Save() CartMemento {
    snapshot := make([]CartItem, len(c.items))
    copy(snapshot, c.items)
    return CartMemento{items: snapshot}
}

func (c *Cart) Restore(m CartMemento) {
    c.items = make([]CartItem, len(m.items))
    copy(c.items, m.items)
}

// Caretaker
type CartHistory struct {
    stack []CartMemento
}

func (h *CartHistory) Push(m CartMemento) { h.stack = append(h.stack, m) }
func (h *CartHistory) Pop() (CartMemento, bool) {
    if len(h.stack) == 0 {
        return CartMemento{}, false
    }
    last := h.stack[len(h.stack)-1]
    h.stack = h.stack[:len(h.stack)-1]
    return last, true
}
```

**Critical:** Deep-copy slices in both `Save()` and `Restore()`. A shallow copy
means the memento shares the backing array with the cart — mutations after
saving will corrupt the snapshot.

**Do NOT use generics here:** Go generics are available in 1.23, but the
educational value is in demonstrating the Originator/Memento/Caretaker roles
explicitly. Generic `Memento[T]` obscures those roles.

**Confidence:** HIGH — confirmed against refactoring.guru Go example and
BigBoxCode article; deep-copy requirement is universal.

---

### Visitor (Issue #7)

**Go idiom:** Go has no method overloading, so the Visitor interface must define
a distinct method per concrete element type. Each element implements `Accept(v Visitor)`.

```go
type OrderExportVisitor interface {
    VisitCSVItem(item *CartItem)
    VisitCSVBundle(bundle *CartBundle)
    VisitCSVDiscount(d *AppliedDiscount)
}

// Element interface
type CartElement interface {
    Accept(v OrderExportVisitor)
}

// Concrete element
func (i *CartItem) Accept(v OrderExportVisitor) {
    v.VisitCSVItem(i)
}

// Concrete visitor
type CSVExporter struct {
    buf strings.Builder
}

func (e *CSVExporter) VisitCSVItem(item *CartItem) {
    e.buf.WriteString(fmt.Sprintf("%s,%d,%.2f\n", item.ProductID, item.Qty, item.Price))
}
```

**Why this naming:** `VisitCSVItem` not `Visit(item *CartItem)` — Go does not
allow two methods with the same name and different parameter types on the same
interface. Each element type needs a distinct method name.

**Tradeoff:** Adding a new element type requires updating the Visitor interface
AND all existing visitor implementations. This is the known Visitor tradeoff —
make it explicit in PR comments for the OOP assessment.

**Confidence:** HIGH — confirmed against refactoring.guru Go Visitor example;
the naming convention limitation is widely documented.

---

### Facade (Issue #8)

**Go idiom:** A `NotificationFacade` struct that holds references to all
notification subsystems (email observer, in-memory observer, log observer) and
exposes simple high-level methods. The facade does NOT implement any of the
subsystem interfaces — it is a new, simpler interface.

```go
type NotificationFacade struct {
    subject     *notification.PriceSubject
    logObs      *notification.LogObserver
    memObs      *notification.InMemoryObserver
}

func NewNotificationFacade(subject *notification.PriceSubject) *NotificationFacade {
    f := &NotificationFacade{subject: subject}
    f.logObs = &notification.LogObserver{UserEmail: "system"}
    f.memObs = &notification.InMemoryObserver{}
    return f
}

// Simple method hiding observer registration complexity
func (f *NotificationFacade) WatchProduct(productID, userEmail string) {
    obs := &notification.LogObserver{UserEmail: userEmail}
    f.subject.Subscribe(productID, obs)
}

func (f *NotificationFacade) GetNotifications() []notification.NotificationRecord {
    return f.memObs.Records
}
```

**Confidence:** HIGH — Facade is the most straightforward pattern in Go;
confirmed against O'Reilly Go Design Patterns and refactoring.guru.

---

### Proxy (Issue #8)

**Go idiom:** The proxy implements the same interface as the real subject. The
logging proxy intercepts method calls, logs them, and delegates to the wrapped
object. Critically, the proxy and the real object must share an interface —
Go has no dynamic proxy generation.

```go
type ProductCatalog interface {
    GetProduct(id string) (models.Product, bool)
    ListProducts() []models.Product
}

// Real subject — already exists in handler
type InMemoryCatalog struct { ... }

// Logging proxy
type LoggingCatalogProxy struct {
    inner  ProductCatalog
    logger *log.Logger
}

func (p *LoggingCatalogProxy) GetProduct(id string) (models.Product, bool) {
    p.logger.Printf("GetProduct called: id=%s", id)
    result, ok := p.inner.GetProduct(id)
    p.logger.Printf("GetProduct result: found=%v", ok)
    return result, ok
}
```

**Important:** The existing `internal/handler/products.go` accesses product data
directly via a map. To introduce a Proxy, that access must be extracted behind
a `ProductCatalog` interface first. This is a prerequisite refactor for Issue #8.

**Confidence:** HIGH — confirmed by O'Reilly and refactoring.guru; the interface
extraction prerequisite is derived from analyzing the existing codebase.

---

### Bridge (Issue #8)

**Go idiom:** Bridge separates an abstraction from its implementation so that
the two can vary independently. In Go, this means two interfaces: one for the
abstraction (e.g., `NotificationSender`) and one for the implementation
(e.g., `OutputFormatter`). The abstraction holds a reference to the implementation.

```go
// Implementation interface — how output is formatted
type OutputFormatter interface {
    Format(msg string) string
}

type JSONFormatter struct{}
func (f *JSONFormatter) Format(msg string) string {
    return fmt.Sprintf(`{"message":%q}`, msg)
}

type PlainTextFormatter struct{}
func (f *PlainTextFormatter) Format(msg string) string { return msg }

// Abstraction — what is sent, not how it is formatted
type NotificationSender struct {
    formatter OutputFormatter
}

func (s *NotificationSender) Send(event string, detail string) string {
    msg := fmt.Sprintf("[%s] %s", event, detail)
    return s.formatter.Format(msg)
}
```

**Why Bridge, not Adapter:** In Adapter, you're adapting existing third-party
code. In Bridge, you design both sides from scratch with variability as the
goal. For this project, the formatter is new code — Bridge is correct.

**Confidence:** MEDIUM — Bridge vs Adapter vs Decorator distinction in Go is
contextual; the domain fit (swappable output format for notifications) is
plausible but the educational line is thin. The key is to make the "two axes
of variation" explicit in PR documentation.

---

## Common Mistakes When Implementing OOP Patterns in Go

### Mistake 1: Trying to Simulate Inheritance for Template Method

**What happens:** Developer embeds a `BaseProcessor` struct and puts the
template logic inside it, expecting the embedded struct's methods to dispatch
to the outer struct's overrides. They do not. `base.ProcessOrder()` calls
`base.Validate()`, not the outer struct's `Validate()`.

**Fix:** The template is always a standalone function or a function that accepts
an interface. The interface is the variability point, not the embedded struct.

### Mistake 2: Channel-Based Iterator

**What happens:** Iterator returns a `chan *Order` and a goroutine sends values
into it. Works, but leaks the goroutine if the consumer breaks early unless
a done channel is also threaded through.

**Fix:** Use `iter.Seq[*Order]` with yield. Early break is handled by yield
returning false. No goroutine needed.

### Mistake 3: Shallow Copy in Memento

**What happens:** `Save()` copies the slice header but not the backing array.
Subsequent mutations to the cart modify the "saved" snapshot.

**Fix:** `copy(snapshot, c.items)` in both `Save()` and `Restore()`. If items
contain pointer fields, those fields need deep-copying too.

### Mistake 4: Visitor Interface with Overloaded Method Names

**What happens:** Developer writes `Visit(item *CartItem)` and `Visit(bundle *CartBundle)`
expecting overloading. Go does not allow two methods with the same name on the
same type.

**Fix:** Use distinct method names: `VisitCartItem`, `VisitCartBundle`. Document
this as a consequence of Go's design in the PR.

### Mistake 5: State as String Constants with Switch

**What happens:** Order state stored as `status string` with a `switch status`
block in every method. This is a state machine but NOT the State pattern.

**Fix:** Each state is a struct implementing the `State` interface. Transitions
are methods on those structs, not cases in a switch.

### Mistake 6: Mediator Becomes God Object

**What happens:** All business logic accumulates in the mediator's `Notify`
method as the system grows.

**Fix:** Mediator handles routing/coordination only. Business logic stays in
components or services. Keep the mediator's switch cases thin — they dispatch,
not compute.

### Mistake 7: Proxy Without Interface Extraction

**What happens:** Proxy wraps a concrete struct, requiring callers to know the
proxy type. The pattern only works transparently if both proxy and real subject
implement a shared interface.

**Fix:** Extract the interface first. The handler uses the interface type, not
the concrete struct. Then the proxy is a drop-in replacement.

---

## Existing Pattern Conventions (Issues #1–3, to maintain consistency)

| Pattern | Location | Go Idiom Used |
|---------|----------|---------------|
| Factory Method | `internal/factory/product_factory.go` | `ProductFactory` interface + `GetFactory(string)` dispatch function |
| Abstract Factory | `internal/factory/brand_factory.go` | Interface returning product families |
| Prototype | `internal/bundle/bundle.go` | `Clone()` method with explicit `copy()` for slice fields |
| Builder | `internal/bundle/builder.go` | Method chaining on `*BundleBuilder`; `Build()` validates and resets |
| Strategy | `internal/discount/strategy.go` | `Strategy` interface with `Apply(float64) float64` |
| Command | `internal/discount/command.go` | `Command` interface with `Execute() float64` and `Undo() float64` |
| Observer | `internal/notification/observer.go` | `PriceObserver` interface; `PriceSubject` holds `map[string][]PriceObserver` |

New patterns should follow the same conventions: interface at top of file,
concrete types below, no exported fields on concrete types unless needed for JSON.

---

## Package Structure for Remaining Issues

| Issue | New Package | Reason |
|-------|-------------|--------|
| #4 | `internal/order` | Created fresh; contains Order struct, MacroCommand, Template |
| #5 | extends `internal/order` | State, Iterator, Chain of Responsibility add to existing Order |
| #6 | `internal/search` + `internal/chat` | Interpreter in search, Mediator in chat |
| #7 | `internal/cart` | Memento and Visitor operate on Cart |
| #8 | `internal/facade` + refactors in `internal/handler` | Facade, Proxy, Bridge wrap existing subsystems |

---

## Sources

- Go official blog — Range Over Function Types: https://go.dev/blog/range-functions (HIGH confidence)
- Refactoring.Guru — Go Design Patterns: https://refactoring.guru/design-patterns/go (HIGH confidence)
- Refactoring.Guru — State in Go: https://refactoring.guru/design-patterns/state/go/example (HIGH confidence)
- Refactoring.Guru — Chain of Responsibility in Go: https://refactoring.guru/design-patterns/chain-of-responsibility/go/example (HIGH confidence)
- Refactoring.Guru — Visitor in Go: https://refactoring.guru/design-patterns/visitor/go/example (HIGH confidence)
- Refactoring.Guru — Memento in Go: https://refactoring.guru/design-patterns/memento/go/example (HIGH confidence)
- Refactoring.Guru — Mediator in Go: https://refactoring.guru/design-patterns/mediator/go/example (HIGH confidence)
- Maurício Linhares — GoF Patterns That Still Make Sense in Go: https://mauricio.github.io/2022/02/07/gof-patterns-in-golang.html (MEDIUM confidence)
- golang-nuts thread — Template Method implementation: https://groups.google.com/g/golang-nuts/c/dzpJ_riiRZ4 (MEDIUM confidence)
- TutorialEdge — Go 1.23 Iterators: https://tutorialedge.net/golang/go-123-iterators-tutorial/ (HIGH confidence)
- Hacking with Go — Interpreter pattern: https://www.hackingwithgo.nl/2023/04/16/design-patterns-in-go-interpreter-making-sense-of-the-world/ (MEDIUM confidence)
- BigBoxCode — Memento Pattern Go: https://bigboxcode.com/design-pattern-memento-pattern-go (MEDIUM confidence)
