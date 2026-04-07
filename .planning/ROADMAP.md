# Roadmap: PawShop

## Overview

Five phases, one per GitHub issue (#4-#8). Each phase introduces specific OOP design patterns applied to a working Go REST API. Phases 1 and 2 are sequential (Phase 2 extends the `internal/order` package created in Phase 1). Phases 3, 4, and 5 are independent and can follow Phase 2 in any order.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Batch Orders** - MacroCommand and Template Method for batch order actions and report generation (#4) (completed 2026-04-07)
- [x] **Phase 2: Order Lifecycle** - State, Iterator, and Chain of Responsibility for order transitions, listing, and validation (#5) (completed 2026-04-07)
- [ ] **Phase 3: Search & Chat** - Interpreter for product search and Mediator for support chat (#6)
- [ ] **Phase 4: Cart & Export** - Memento for undo and Visitor for cart export (#7)
- [ ] **Phase 5: Notifications & Logging** - Facade for notifications, Proxy for logging, Bridge for output formatting (#8)

## Phase Details

### Phase 1: Batch Orders
**Goal**: Callers can batch-confirm or batch-reject orders and generate daily or weekly reports
**Depends on**: Nothing (first phase)
**Requirements**: BATCH-01, BATCH-02, BATCH-03, BATCH-04, BATCH-05, BATCH-06, BATCH-07, BATCH-08, BATCH-09, BATCH-10, BATCH-11
**Success Criteria** (what must be TRUE):
  1. POST /api/orders/batch with order_ids and action "confirm" or "reject" changes each order's status via MacroCommand
  2. GET /api/reports/daily returns a short daily summary; GET /api/reports/weekly returns a detailed weekly report
  3. MacroCommand unit tests pass for multi-command execution and error behavior
  4. Template Method unit tests pass for both report types
  5. HTTP handler tests pass for batch and report endpoints
**Plans:** 2/2 plans complete
Plans:
- [x] 01-01-PLAN.md — Order domain: MacroCommand + Template Method + unit tests
- [x] 01-02-PLAN.md — HTTP handler, handler tests, and route registration

### Phase 2: Order Lifecycle
**Goal**: Orders move through a defined lifecycle, can be listed and filtered, and are validated before creation
**Depends on**: Phase 1
**Requirements**: LIFE-01, LIFE-02, LIFE-03, LIFE-04, LIFE-05, LIFE-06, LIFE-07, LIFE-08, LIFE-09, LIFE-10, LIFE-11, LIFE-12, LIFE-13, LIFE-14
**Success Criteria** (what must be TRUE):
  1. PATCH /api/orders/{id}/status with action "next" advances the order state; "cancel" cancels it from allowed states; illegal transitions return errors
  2. GET /api/orders returns all orders; GET /api/orders?status=X returns only orders in that state
  3. POST /api/orders runs through StockValidator, AddressValidator, and PaymentValidator — any failure returns a descriptive error
  4. State, Iterator, and Chain of Responsibility unit and handler tests all pass
**Plans:** 3/3 plans complete
Plans:
- [x] 02-01-PLAN.md — State pattern (order lifecycle transitions) + Iterator (collection traversal) + unit tests
- [x] 02-02-PLAN.md — Chain of Responsibility (order validation) + unit tests
- [x] 02-03-PLAN.md — HTTP handlers (PATCH/GET/POST) + handler tests + route registration

### Phase 3: Search & Chat
**Goal**: Users can search products by structured query and exchange messages through a support chat
**Depends on**: Phase 2
**Requirements**: SRCH-01, SRCH-02, SRCH-03, SRCH-04, SRCH-05, SRCH-06, SRCH-07, SRCH-08, SRCH-09, SRCH-10, SRCH-11
**Success Criteria** (what must be TRUE):
  1. GET /api/products/search?q=brand:X or ?q=price:<N or ?q=category:X returns filtered results via Interpreter
  2. POST /api/chat/send routes messages through the Mediator; participants never hold direct references to each other
  3. GET /api/chat/history?participant=X returns messages for that participant
  4. Interpreter and Mediator unit and handler tests all pass
**Plans**: TBD

### Phase 4: Cart & Export
**Goal**: Users can manage a cart with undo support and export it in JSON or text format
**Depends on**: Phase 2
**Requirements**: CART-01, CART-02, CART-03, CART-04, CART-05, CART-06, CART-07, CART-08, CART-09, CART-10, CART-11, CART-12, CART-13
**Success Criteria** (what must be TRUE):
  1. POST /api/cart/add and POST /api/cart/remove modify the cart; POST /api/cart/undo restores the previous state via Memento
  2. GET /api/cart returns current cart contents
  3. GET /api/cart/export?format=json returns a JSON export; ?format=text returns a human-readable receipt
  4. Memento and Visitor unit and handler tests all pass
**Plans**: TBD

### Phase 5: Notifications & Logging
**Goal**: Notifications are sent through a unified facade, logs are filtered via a proxy, and output is formatted through a bridge
**Depends on**: Phase 2
**Requirements**: NOTF-01, NOTF-02, NOTF-03, NOTF-04, NOTF-05, NOTF-06, NOTF-07, NOTF-08, NOTF-09, NOTF-10, NOTF-11, NOTF-12, NOTF-13, NOTF-14, NOTF-15
**Success Criteria** (what must be TRUE):
  1. POST /api/notifications/send triggers NotifyUser via Facade, which internally uses ConsoleNotifier and FileNotifier
  2. GET /api/logs?level=X returns log entries filtered by level via LoggerProxy; GET /api/logs/stats returns counts by level
  3. All formatter+writer Bridge combinations (Text/JSON x Console/File) produce correct output
  4. Facade, Proxy, and Bridge unit and handler tests all pass
**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in order: 1 -> 2 -> 3/4/5 (3, 4, 5 independent after Phase 2)

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Batch Orders | 2/2 | Complete    | 2026-04-07 |
| 2. Order Lifecycle | 3/3 | Complete    | 2026-04-07 |
| 3. Search & Chat | 0/TBD | Not started | - |
| 4. Cart & Export | 0/TBD | Not started | - |
| 5. Notifications & Logging | 0/TBD | Not started | - |
