# Domain Pitfalls: Go Design Patterns

**Domain:** OOP design patterns implemented in Go (composition-over-inheritance model)
**Project:** PawShop — educational dog food store
**Researched:** 2026-04-07

---

## Critical Pitfalls

Mistakes that cause rewrites, broken pattern intent, or hard-to-debug runtime panics.

---

### Pitfall 1: Template Method — Method Dispatch Breaks Without Interface Self-Reference

**What goes wrong:** Go has no abstract classes. The naive approach embeds a "base" struct and calls step methods directly on it. The base struct's methods resolve to themselves rather than to the overriding concrete type, so the template method never calls the specialised steps — the algorithm runs the base version every time.

**Why it happens:** In Java, `this` inside a base class always refers to the concrete subclass at runtime. In Go, embedding does not do this. When `Otp.genAndSendOTP()` calls `self.genRandomOTP()`, "self" is the embedded `Otp` value, not the outer `Sms` or `Email` struct.

**The fix:** Store the interface reference inside the template struct and assign `self` at construction time:
```go
type OrderProcessor struct {
    impl IOrderProcessor  // assigned to the concrete type in its constructor
}
func (p *OrderProcessor) Process() {
    p.impl.Validate()
    p.impl.Execute()
    p.impl.Notify()
}
```
Without this pattern, calling `p.Validate()` inside the template always resolves to the base no-op.

**Consequences:** Pattern appears to work in tests that only instantiate the base; breaks silently when subclasses are introduced. The bug is invisible until a second concrete implementation is added.

**Warning signs:** All implementations produce identical output regardless of which concrete type is used. Step methods on the concrete type are never called during tracing/debugging.

**Prevention:**
1. Always define a step interface (`IOrderProcessor`) and store it as a field in the template struct.
2. Each concrete constructor must assign `self` to that field: `impl: s` where `s` is `*ShipOrder`.
3. Write a test with two concrete implementations that produce different step output.

**Phase:** Issue #4 (Batch order actions — Template Method)

---

### Pitfall 2: Memento — Shallow Copy Saves a Reference, Not a Snapshot

**What goes wrong:** A Memento captures `state := cart.items` (slice assignment). Since slices are reference types, the memento holds a pointer to the same backing array. Subsequent mutations to the cart corrupt the snapshot, making undo produce wrong results.

**Why it happens:** Go slice assignment copies the slice header (pointer + length + capacity) but not the underlying array. Maps have identical behaviour. Both are reference semantics, not value semantics.

**The fix:** Always deep-copy slices and maps when creating a snapshot:
```go
func (c *Cart) Save() *CartMemento {
    items := make([]CartItem, len(c.items))
    copy(items, c.items)
    return &CartMemento{items: items}
}
```
For nested pointer fields, copy each pointer target individually or use a serialise/deserialise round-trip (JSON marshal/unmarshal is simple and testable without external deps).

**Consequences:** Undo appears to work on a simple happy-path test. Fails only when a second mutation happens before undo is invoked — the kind of bug that passes unit tests and breaks acceptance tests.

**Warning signs:** After `AddItem → AddItem → Undo`, the cart still contains both items instead of one.

**Prevention:**
1. Every `Save()` method must explicitly copy each slice and map field.
2. Write a table-driven test: Add, Add, Undo, verify count == 1; Add, Add, Undo, Undo, verify count == 0.
3. Consider adding a comment `// deep copy — not assignment` above each copy block.

**Phase:** Issue #7 (Cart with undo — Memento)

---

### Pitfall 3: Visitor — Interface Explosion When Element Types Grow

**What goes wrong:** The Visitor interface must declare a `Visit*` method for every element type. Adding a new product type (e.g., `Supplement`) requires updating every existing visitor implementation. If any visitor is missed, the code fails to compile — which is actually Go's best-case outcome. More often, developers add a catch-all empty method, defeating the pattern.

**Why it happens:** Go's lack of method overloading means each concrete element needs its own visit method signature. With 3 product types (DryFood, WetFood, Treat) the visitor interface already has 3 methods; adding export formats multiplies quickly.

**The fix for this project:** Keep the element set fixed (the 3 existing product types). The visitor interface should have exactly 3 methods and never grow mid-implementation. Define it once, test it, lock it.

**Consequences:** Blank `VisitSupplement(s *Supplement) {}` stubs propagate across every visitor, hiding unimplemented behaviour and breaking the pattern's guarantee that every visitor handles every type.

**Warning signs:** Any visitor implementation has methods with empty bodies that were added "just to compile".

**Prevention:**
1. Define the full element set before writing the Visitor interface — no additions mid-feature.
2. Avoid a default/fallback `Visit(e Element)` method; force compilation errors for missing implementations.
3. Document the closed element set explicitly in a comment on the interface.

**Phase:** Issue #7 (Order export — Visitor)

---

### Pitfall 4: Iterator — Channel-Based Iterator Leaks Goroutines

**What goes wrong:** Implementing Iterator with a goroutine that sends values on a channel is idiomatic-looking but has a critical leak: if the consumer stops iterating early (e.g., finds what it needs and returns), the producer goroutine is permanently blocked on `ch <- item` and never exits.

**Why it happens:** The goroutine has no way to know iteration was abandoned. It tries to send the next item, blocks forever, and is never garbage collected.

**The fix for Go 1.23:** Use the stateful cursor (Next()/Value() method) approach or a simple index-based struct. Go 1.22+ range-over-functions experiment exists but is not stable in 1.23. For this educational project, the `Next() bool` + `Value()` approach matches stdlib patterns (`sql.Rows`, `bufio.Scanner`) and has no leak risk:
```go
type OrderIterator struct {
    orders []*Order
    index  int
}
func (it *OrderIterator) HasNext() bool { return it.index < len(it.orders) }
func (it *OrderIterator) Next() *Order  { o := it.orders[it.index]; it.index++; return o }
```

**Consequences:** Tests pass (tests exhaust the iterator). Production use with early-exit creates goroutine accumulation — measurable with `runtime.NumGoroutine()` but invisible in normal testing.

**Warning signs:** `runtime.NumGoroutine()` grows with each request; `go tool pprof` goroutine profile shows blocked goroutines on channel sends in the iterator package.

**Prevention:**
1. Do not use channels for iterators in this project. Use index-based struct.
2. If channels must be used, pass a `done` channel or `context.Context` to allow producer shutdown.
3. Add a `runtime.NumGoroutine()` assertion in tests that confirm goroutine count does not grow across repeated iteration.

**Phase:** Issue #5 (Order lifecycle — Iterator)

---

## Moderate Pitfalls

---

### Pitfall 5: State — Circular Reference Initialization Order

**What goes wrong:** States hold a pointer back to the context (`*OrderContext`), and the context holds pointers to all states. If states are initialised lazily (created on first transition), the first method call may encounter a nil state pointer because the initial state was never assigned.

**Why it happens:** In some implementations, states are allocated inline in transition methods. If two states depend on each other during transitions, one may be nil at the moment a transition fires.

**The fix:** Create all concrete state instances inside the context constructor and assign `initialState` before returning:
```go
func NewOrderContext() *OrderContext {
    c := &OrderContext{}
    c.pendingState = &PendingState{ctx: c}
    c.shippedState = &ShippedState{ctx: c}
    // ... all states
    c.setState(c.pendingState)
    return c
}
```
Go's GC handles the circular pointer graph correctly — this is not a memory leak concern, only an initialisation order concern.

**Consequences:** Runtime nil-pointer panic on the first state transition in production, but only if the initial state is set anywhere other than the constructor.

**Warning signs:** `nil pointer dereference` panics in `(*OrderContext).setState` or any state method. Tests pass when run individually but panic in integration sequences.

**Prevention:**
1. Set all state instances and the initial state inside the constructor, never elsewhere.
2. Validate in tests: construct an `OrderContext` and immediately call `GetState()` — it must return non-nil.

**Phase:** Issue #5 (Order lifecycle — State)

---

### Pitfall 6: Chain of Responsibility — Silent Request Drops

**What goes wrong:** When a request reaches the end of the chain without being handled (no handler claims it), it disappears silently. Callers receive no error, no signal, and no indication the request was unprocessed.

**Why it happens:** The idiomatic Go implementation calls `next.Handle(req)` if the current handler passes; if `next` is nil, the method simply returns. There is no enforcement that someone must handle the request.

**The fix:** Terminate the chain with a "catch-all" handler that returns a structured error or a terminal response:
```go
type UnhandledHandler struct{}
func (h *UnhandledHandler) Handle(req *OrderRequest) error {
    return fmt.Errorf("no handler accepted order request: %v", req.Type)
}
```
Or: the last real handler returns an explicit error if it cannot handle the request.

**Consequences:** Order validation requests silently no-op when no validator in the chain claims responsibility. HTTP handler returns 200 with empty body instead of 422.

**Warning signs:** A request type that should be rejected passes through all handlers and produces a 200 response with no action.

**Prevention:**
1. Always define a terminal "catch-all" or "default" handler at chain construction.
2. Write a test: send an unknown request type, assert a non-nil error is returned.

**Phase:** Issue #5 (Order lifecycle — Chain of Responsibility)

---

### Pitfall 7: Interpreter — Missing Recursion Depth Guard (Stack Overflow)

**What goes wrong:** A recursive-descent Interpreter with no depth limit panics with a stack overflow on deeply nested or malformed query expressions. Go does not perform tail-call optimisation.

**Why it happens:** Each recursive call to `Parse()` or `Interpret()` adds a stack frame. On a pathological input like `((((((((query))))))))`, depth is proportional to input length. The Go runtime will crash if the stack cannot grow further.

**Consequences:** An attacker or a test with a complex query string can crash the server with a trivially crafted input. CVE-2024-34155 demonstrated exactly this in `go/parser` itself.

**Warning signs:** Server panics on long or deeply-nested search query strings. Crash reproduced with `strings.Repeat("(", 1000)` as input.

**Prevention:**
1. Add a `maxDepth int` parameter to recursive parse functions; return an error when exceeded.
2. Set a reasonable limit (e.g., 32 levels of nesting) appropriate for a product search query.
3. Write a fuzz-like test: pass a string of 100 nested brackets and assert it returns an error, not a panic.
4. Use `recover()` in the HTTP handler as a last-resort guard, but do not rely on it as the primary protection.

**Phase:** Issue #6 (Product search — Interpreter)

---

### Pitfall 8: Proxy — Interface-Only Proxies Require Interface Refactoring

**What goes wrong:** A Proxy can only wrap a concrete type if that type implements an interface. If the existing notification or logging implementation is a concrete struct (not hiding behind an interface), adding a Proxy requires retrofitting the calling code to use the interface — which touches multiple files.

**Why it happens:** Go proxies have no dynamic proxy generation (unlike Java's `Proxy.newProxyInstance`). Go lacks the dynamic dispatch hook needed to intercept arbitrary method calls on a concrete type.

**The fix:** Ensure the component being proxied (`NotificationService`, logging target) is already exposed as an interface before the proxy is written. For this project, plan the interface in the same phase the proxy is introduced.

**Consequences:** Adding a LoggingProxy to an existing concrete `NotificationSender` requires: define an interface, change all call sites to use the interface, then write the proxy. Mid-phase refactoring breaks the commit-per-issue workflow.

**Warning signs:** The proxy struct cannot be substituted at call sites because call sites hold `*ConcreteType`, not `InterfaceType`.

**Prevention:**
1. Issue #8 should define the notification/logging interface before implementing the Proxy.
2. Review Issue #3's notification package: if `PriceSubject` methods are called directly (not via interface), add an interface wrapper in Issue #8's first step.

**Phase:** Issue #8 (Logging proxy — Proxy)

---

## Minor Pitfalls

---

### Pitfall 9: MacroCommand — Unbounded History Growth

**What goes wrong:** The existing `CommandHistory` in `internal/discount/command.go` appends commands indefinitely. In a long-running server, applying thousands of batch order commands fills the history slice without bound.

**Why it happens:** The `history []Command` slice is never capped. Each `Execute()` appends; nothing prunes old entries.

**Prevention:** Cap the history at a reasonable maximum (e.g., 100 entries) or clear it after a bulk operation completes. For an educational project this is low-risk but worth noting in the MacroCommand implementation: the `MacroCommand` itself is a single entry in history, not N entries — so the issue is at the `MacroCommand` aggregation level.

**Phase:** Issue #4 (Batch order actions — MacroCommand)

---

### Pitfall 10: Facade — Hiding Errors Inside the Facade

**What goes wrong:** A Facade simplifies a complex subsystem by providing a single method. Developers sometimes suppress or swallow errors from subsystem calls because "the facade should be simple." Callers cannot distinguish between success and subsystem failure.

**Prevention:** Facade methods must return errors that propagate from subsystem calls. Simplicity means fewer methods, not fewer error paths. Use `errors.Join` (Go 1.20+) to aggregate multiple subsystem errors into one return.

**Phase:** Issue #8 (Notifications facade — Facade)

---

### Pitfall 11: Bridge — Abstraction and Implementor Interface Confusion

**What goes wrong:** When defining Bridge for output formatting (e.g., JSON vs plain text output), developers conflate the abstraction layer with the implementor layer. Both end up implementing the same interface, and the Bridge indirection adds no actual decoupling.

**Prevention:** The abstraction defines _what_ operation (e.g., "send notification"), the implementor defines _how_ the output is formatted (e.g., JSON encoder, text encoder). These are distinct responsibilities. Name them explicitly: `NotificationSender` (abstraction) and `OutputFormatter` (implementor).

**Phase:** Issue #8 (Output bridge — Bridge)

---

### Pitfall 12: Mediator — Circular Import If Components Are in Separate Packages

**What goes wrong:** If the Mediator and its components are split across packages, Go's no-circular-import rule forces careful package design. A chat message component that imports the mediator package, and a mediator that imports the component package, causes a compile error.

**Prevention:** Place the Mediator interface and all component types in the same package (`internal/chat` or `internal/support`). Define the Mediator as an interface in that package so components depend on the interface, not the concrete mediator struct.

**Phase:** Issue #6 (Support chat — Mediator)

---

## Phase-Specific Warning Map

| Issue | Pattern | Likely Pitfall | Mitigation |
|-------|---------|---------------|------------|
| #4 | Template Method | Method dispatch resolves to base, not concrete type | Store interface self-reference in template struct |
| #4 | MacroCommand | History grows unbounded | Cap history; aggregate sub-commands into single history entry |
| #5 | State | Nil state on first transition if lazily initialised | Initialise all states in constructor |
| #5 | Iterator | Goroutine leak from channel-based iterator | Use index cursor pattern, not channels |
| #5 | Chain of Responsibility | Silent drop of unhandled requests | Terminal catch-all handler required |
| #6 | Interpreter | Stack overflow on deeply nested queries | Depth guard on all recursive parse calls |
| #6 | Mediator | Circular import if mediator and components in separate packages | Co-locate in single package |
| #7 | Memento | Slice/map shallow copy breaks snapshot semantics | Explicit `make` + `copy` in every Save() |
| #7 | Visitor | Interface explosion when element set is not fixed | Lock element set before writing Visitor interface |
| #8 | Proxy | Concrete types cannot be proxied; need interface refactoring | Define interface in same phase as Proxy |
| #8 | Facade | Errors swallowed for "simplicity" | Facade methods must propagate errors |
| #8 | Bridge | Abstraction and implementor conflated | Name each layer with distinct responsibility |

---

## Sources

- Refactoring.Guru — Template Method in Go: https://refactoring.guru/design-patterns/template-method/go/example
- Refactoring.Guru — State in Go: https://refactoring.guru/design-patterns/state/go/example
- Refactoring.Guru — Memento in Go: https://refactoring.guru/design-patterns/memento/go/example
- Refactoring.Guru — Chain of Responsibility in Go: https://refactoring.guru/design-patterns/chain-of-responsibility/go/example
- Hacking with Go — Chain of Responsibility pitfalls: https://www.hackingwithgo.nl/2023/04/30/design-patterns-in-go-chain-of-responsibility-there-is-more-than-one-way-to-do-it/
- Mauricio Linhares — GoF patterns in Go: https://mauricio.github.io/2022/02/07/gof-patterns-in-golang.html
- YourBasic — Iterator in Go: https://yourbasic.org/golang/iterator-generator-pattern/
- Goroutine leak from channel iterators: https://groups.google.com/g/golang-nuts/c/pxfhKGqHNv0
- Go CVE-2024-34155 — recursion depth in go/parser: https://github.com/golang/go/issues (fixed in Go 1.22.7/1.23.1)
- Deep copy pitfalls in Go: https://leapcell.io/blog/deep-copy-in-golang-techniques-and-best-practices
- Go anti-patterns: https://programmerscareer.com/golang-anti-patterns/
- Nil pointer dereference in Go interfaces: https://oneuptime.com/blog/post/2026-01-23-go-nil-pointer-panics/view
- Proxy limitations in Go (no inheritance): https://mauricio.github.io/2022/02/07/gof-patterns-in-golang.html
