# Requirements: PawShop

**Defined:** 2026-04-07
**Core Value:** Each feature clearly demonstrates its assigned design patterns through working, tested Go code

## v1 Requirements

Requirements for issues #4-#8. Each maps to one roadmap phase.

### Batch Orders (#4)

- [ ] **BATCH-01**: MacroCommand composes multiple OrderCommands and executes them sequentially
- [ ] **BATCH-02**: ConfirmOrderCommand changes order status to "confirmed"
- [ ] **BATCH-03**: RejectOrderCommand changes order status to "rejected"
- [ ] **BATCH-04**: POST /api/orders/batch accepts order_ids and action ("confirm"|"reject"), applies MacroCommand
- [ ] **BATCH-05**: Template Method defines abstract report algorithm with Header/Body/Footer steps
- [ ] **BATCH-06**: DailyReportGenerator produces short daily summary
- [ ] **BATCH-07**: WeeklyReportGenerator produces detailed weekly format
- [ ] **BATCH-08**: GET /api/reports/{type} returns generated report (type = "daily"|"weekly")
- [ ] **BATCH-09**: Unit tests for MacroCommand (multi-command execution, error behavior)
- [ ] **BATCH-10**: Unit tests for Template Method (both report types)
- [ ] **BATCH-11**: HTTP handler tests for batch and report endpoints

### Order Lifecycle (#5)

- [ ] **LIFE-01**: State pattern controls order transitions: New → Confirmed → Shipped → Delivered
- [ ] **LIFE-02**: Each state enforces allowed transitions and returns errors on illegal ones
- [ ] **LIFE-03**: Cancel is available from appropriate states
- [ ] **LIFE-04**: PATCH /api/orders/{id}/status with action "next"|"cancel"
- [ ] **LIFE-05**: Iterator provides HasNext/Next over order collection
- [ ] **LIFE-06**: Filtered iterator supports filtering by status
- [ ] **LIFE-07**: GET /api/orders lists orders, with optional ?status= query filter
- [ ] **LIFE-08**: Chain of Responsibility validates orders: StockValidator → AddressValidator → PaymentValidator
- [ ] **LIFE-09**: Each validator can fail independently with descriptive error
- [ ] **LIFE-10**: POST /api/orders creates order through validation chain
- [ ] **LIFE-11**: Unit tests for State (allowed/forbidden transitions)
- [ ] **LIFE-12**: Unit tests for Iterator (iteration, filtered iteration)
- [ ] **LIFE-13**: Unit tests for Chain of Responsibility (success, failure at each step)
- [ ] **LIFE-14**: HTTP handler tests for all order endpoints

### Search & Chat (#6)

- [ ] **SRCH-01**: Interpreter parses text queries: brand:X, price:<N, category:X
- [ ] **SRCH-02**: AndExpression combines two expressions with logical AND
- [ ] **SRCH-03**: Parse function converts query string to Expression tree
- [ ] **SRCH-04**: GET /api/products/search?q= filters products using interpreter
- [ ] **SRCH-05**: Mediator coordinates chat between Customer and Manager participants
- [ ] **SRCH-06**: Participants communicate only through mediator (no direct references)
- [ ] **SRCH-07**: POST /api/chat/send sends message via mediator
- [ ] **SRCH-08**: GET /api/chat/history?participant= returns message history
- [ ] **SRCH-09**: Unit tests for Interpreter (each expression type, parsing, combined queries)
- [ ] **SRCH-10**: Unit tests for Mediator (message routing, participant isolation)
- [ ] **SRCH-11**: HTTP handler tests for search and chat endpoints

### Cart & Export (#7)

- [ ] **CART-01**: Cart supports add, remove, and update quantity operations
- [ ] **CART-02**: Memento saves cart state (deep copy) before each mutation
- [ ] **CART-03**: CartHistory maintains stack of mementos for undo
- [ ] **CART-04**: POST /api/cart/undo restores previous cart state
- [ ] **CART-05**: POST /api/cart/add, POST /api/cart/remove endpoints
- [ ] **CART-06**: GET /api/cart returns current cart state
- [ ] **CART-07**: Visitor pattern with OrderVisitor interface (VisitItem, VisitCart methods)
- [ ] **CART-08**: JSONExportVisitor generates JSON export
- [ ] **CART-09**: TextReceiptVisitor generates human-readable receipt
- [ ] **CART-10**: GET /api/cart/export?format=json|text returns formatted export
- [ ] **CART-11**: Unit tests for Memento (save/restore, undo after add, undo after remove)
- [ ] **CART-12**: Unit tests for Visitor (JSON export, text receipt, empty cart)
- [ ] **CART-13**: HTTP handler tests for cart endpoints

### Notifications & Logging (#8)

- [ ] **NOTF-01**: Facade provides NotifyUser and NotifyOrderStatusChanged methods
- [ ] **NOTF-02**: Facade internally uses ConsoleNotifier and FileNotifier subsystems
- [ ] **NOTF-03**: POST /api/notifications/send sends notification via facade
- [ ] **NOTF-04**: Proxy wraps Logger interface with lazy initialization
- [ ] **NOTF-05**: LoggerProxy counts log entries (GetLogCount)
- [ ] **NOTF-06**: LoggerProxy filters by log level (info, warn, error)
- [ ] **NOTF-07**: GET /api/logs?level= returns filtered log entries
- [ ] **NOTF-08**: GET /api/logs/stats returns log counts by level
- [ ] **NOTF-09**: Bridge separates Formatter abstraction from OutputWriter implementation
- [ ] **NOTF-10**: TextFormatter and JSONFormatter as abstraction variants
- [ ] **NOTF-11**: ConsoleWriter and FileWriter as implementor variants
- [ ] **NOTF-12**: Unit tests for Facade (multi-channel notification)
- [ ] **NOTF-13**: Unit tests for Proxy (lazy init, counter, level filtering)
- [ ] **NOTF-14**: Unit tests for Bridge (all formatter+writer combinations)
- [ ] **NOTF-15**: HTTP handler tests for notification and logging endpoints

## v2 Requirements

Deferred to future. Not in current roadmap.

### Extended Patterns

- **EXT-01**: Decorator pattern for product pricing
- **EXT-02**: Flyweight for shared product attributes
- **EXT-03**: Adapter for external API integration

## Out of Scope

| Feature | Reason |
|---------|--------|
| Database persistence | Educational project — in-memory sufficient |
| Authentication/sessions | Not needed for pattern demonstration |
| Frontend SPA | Static landing page sufficient |
| External dependencies | Constraint: stdlib only |
| Deployment/CI | Local development only |
| Real concurrency patterns | Goroutine patterns not in OOP course scope |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| BATCH-01 | Phase 1 | Pending |
| BATCH-02 | Phase 1 | Pending |
| BATCH-03 | Phase 1 | Pending |
| BATCH-04 | Phase 1 | Pending |
| BATCH-05 | Phase 1 | Pending |
| BATCH-06 | Phase 1 | Pending |
| BATCH-07 | Phase 1 | Pending |
| BATCH-08 | Phase 1 | Pending |
| BATCH-09 | Phase 1 | Pending |
| BATCH-10 | Phase 1 | Pending |
| BATCH-11 | Phase 1 | Pending |
| LIFE-01 | Phase 2 | Pending |
| LIFE-02 | Phase 2 | Pending |
| LIFE-03 | Phase 2 | Pending |
| LIFE-04 | Phase 2 | Pending |
| LIFE-05 | Phase 2 | Pending |
| LIFE-06 | Phase 2 | Pending |
| LIFE-07 | Phase 2 | Pending |
| LIFE-08 | Phase 2 | Pending |
| LIFE-09 | Phase 2 | Pending |
| LIFE-10 | Phase 2 | Pending |
| LIFE-11 | Phase 2 | Pending |
| LIFE-12 | Phase 2 | Pending |
| LIFE-13 | Phase 2 | Pending |
| LIFE-14 | Phase 2 | Pending |
| SRCH-01 | Phase 3 | Pending |
| SRCH-02 | Phase 3 | Pending |
| SRCH-03 | Phase 3 | Pending |
| SRCH-04 | Phase 3 | Pending |
| SRCH-05 | Phase 3 | Pending |
| SRCH-06 | Phase 3 | Pending |
| SRCH-07 | Phase 3 | Pending |
| SRCH-08 | Phase 3 | Pending |
| SRCH-09 | Phase 3 | Pending |
| SRCH-10 | Phase 3 | Pending |
| SRCH-11 | Phase 3 | Pending |
| CART-01 | Phase 4 | Pending |
| CART-02 | Phase 4 | Pending |
| CART-03 | Phase 4 | Pending |
| CART-04 | Phase 4 | Pending |
| CART-05 | Phase 4 | Pending |
| CART-06 | Phase 4 | Pending |
| CART-07 | Phase 4 | Pending |
| CART-08 | Phase 4 | Pending |
| CART-09 | Phase 4 | Pending |
| CART-10 | Phase 4 | Pending |
| CART-11 | Phase 4 | Pending |
| CART-12 | Phase 4 | Pending |
| CART-13 | Phase 4 | Pending |
| NOTF-01 | Phase 5 | Pending |
| NOTF-02 | Phase 5 | Pending |
| NOTF-03 | Phase 5 | Pending |
| NOTF-04 | Phase 5 | Pending |
| NOTF-05 | Phase 5 | Pending |
| NOTF-06 | Phase 5 | Pending |
| NOTF-07 | Phase 5 | Pending |
| NOTF-08 | Phase 5 | Pending |
| NOTF-09 | Phase 5 | Pending |
| NOTF-10 | Phase 5 | Pending |
| NOTF-11 | Phase 5 | Pending |
| NOTF-12 | Phase 5 | Pending |
| NOTF-13 | Phase 5 | Pending |
| NOTF-14 | Phase 5 | Pending |
| NOTF-15 | Phase 5 | Pending |

**Coverage:**
- v1 requirements: 64 total
- Mapped to phases: 64
- Unmapped: 0 ✓

---
*Requirements defined: 2026-04-07*
*Last updated: 2026-04-07 — traceability expanded to individual requirements after roadmap creation*
