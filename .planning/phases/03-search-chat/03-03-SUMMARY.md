---
phase: 03-search-chat
plan: "03"
subsystem: handler
tags: [search, chat, http, interpreter, mediator]
dependency_graph:
  requires: [03-01, 03-02]
  provides: [search-http-endpoint, chat-http-endpoint]
  affects: [cmd/server/main.go, internal/handler]
tech_stack:
  added: []
  patterns: [Interpreter (search filtering via handlers), Mediator (chat routing via handlers)]
key_files:
  created:
    - internal/handler/search.go
    - internal/handler/chat.go
    - internal/handler/search_test.go
    - internal/handler/chat_test.go
  modified:
    - cmd/server/main.go
decisions:
  - name: "Empty matches returns [] not null"
    rationale: "Nil slice marshals to JSON null; initialized empty slice returns [] for consistent API clients"
  - name: "HandleChat routes by path suffix"
    rationale: "Mirrors HandleOrders pattern using strings.HasSuffix for /send and /history sub-paths under /api/chat/"
metrics:
  duration: "~15 minutes"
  completed: "2026-04-07T22:06:47Z"
  tasks_completed: 3
  files_changed: 5
---

# Phase 03 Plan 03: HTTP Handlers for Search and Chat Summary

HTTP handlers wiring Interpreter (search) and Mediator (chat) domain packages to REST endpoints, with full httptest coverage and route registration.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Search HTTP handler | 8b84e2e | internal/handler/search.go |
| 2 | Chat HTTP handler | 8e06517 | internal/handler/chat.go |
| 3 | Handler tests + route registration | 79f5fdb | internal/handler/search_test.go, internal/handler/chat_test.go, cmd/server/main.go |

## What Was Built

**SearchHandler** (`internal/handler/search.go`): GET `/api/products/search?q=<query>` endpoint. Delegates query parsing to `search.Parse()`, builds the product catalog from `RoyalCaninFactory` and `AcanaFactory`, converts each product to `search.ProductData`, filters via `expr.Interpret()`, and returns matching products as `[]models.ProductResponse` JSON.

**ChatHandler** (`internal/handler/chat.go`): POST `/api/chat/send` and GET `/api/chat/history?participant=` endpoints. Constructor pre-registers `customer1` and `manager1` participants via `SupportChatMediator`. `HandleChat` router dispatches by path suffix and method.

**Route registration** (`cmd/server/main.go`): `/api/products/search` registered before `/api/products/` (stdlib mux uses longest-prefix matching, so the more specific path wins). `/api/chat/` registered as prefix route handled by `HandleChat`.

## Test Coverage

**search_test.go** (7 tests):
- `TestHandleSearchBrand` — brand:Royal returns 3 Royal Canin products
- `TestHandleSearchPrice` — price:<500 returns 2 treats (RC Dental Sticks 280, Acana Crunchy Biscuits 350)
- `TestHandleSearchCategory` — category:dry returns 2 dry food products
- `TestHandleSearchCombined` — brand:Royal AND price:<500 returns 1 product (RC Dental Sticks)
- `TestHandleSearchMissingQuery` — missing q param returns 400
- `TestHandleSearchInvalidQuery` — unknown format returns 400
- `TestHandleSearchMethodNotAllowed` — POST returns 405

**chat_test.go** (5 tests):
- `TestHandleSend` — POST with valid body returns 200 with status "sent"
- `TestHandleHistory` — after send, GET history returns 1 message
- `TestHandleSendMissingFields` — missing to/message returns 400
- `TestHandleHistoryMissingParam` — missing participant param returns 400
- `TestHandleSendMethodNotAllowed` — GET on send endpoint returns 405

## Verification

```
go test ./... -count=1   → all packages pass
go vet ./...             → no issues
go build ./...           → builds cleanly
```

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None. All handlers are fully wired to domain packages with live data from factories and mediator.

## Threat Surface Scan

No new surface beyond what the plan's threat model covers:
- T-03-05: `/api/products/search?q=` — mitigated by `search.Parse()` depth guard returning 400 on error
- T-03-06: `/api/chat/send` from field spoofing — accepted (educational project, no auth)
- T-03-07: `/api/chat/history` info disclosure — accepted (no PII in demo data)

## Self-Check: PASSED

Files exist:
- internal/handler/search.go: FOUND
- internal/handler/chat.go: FOUND
- internal/handler/search_test.go: FOUND
- internal/handler/chat_test.go: FOUND

Commits exist:
- 8b84e2e: FOUND
- 8e06517: FOUND
- 79f5fdb: FOUND
