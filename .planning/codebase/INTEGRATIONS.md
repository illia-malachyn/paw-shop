# External Integrations

**Analysis Date:** 2026-04-07

## APIs & External Services

**Not Detected:**
- No external APIs or third-party services integrated
- All data is generated in-memory and not persisted
- No authentication provider integrations

## Data Storage

**Databases:**
- None - Application uses in-memory data only
- No database client configured
- No ORM or persistence layer

**File Storage:**
- Local filesystem only
- Static files served from `static/` directory via `http.FileServer`
- No cloud storage integration

**Caching:**
- In-memory only (ephemeral)
- No dedicated caching service

## Authentication & Identity

**Auth Provider:**
- None - No authentication implemented
- Application is public/open access
- No user session management

## Monitoring & Observability

**Error Tracking:**
- None - No error tracking service (Sentry, Rollbar, etc.)

**Logs:**
- Standard output only
- `log.Fatalf()` for startup errors in `cmd/server/main.go`
- `fmt.Printf()` for informational messages
- `notification.LogObserver` in `internal/notification/observer.go` logs price changes to console

## CI/CD & Deployment

**Hosting:**
- Not specified - Standard HTTP server listening on `localhost:8080`
- Can be deployed to any environment with Go 1.23

**CI Pipeline:**
- Not configured - No CI/CD files detected (.github/workflows, .gitlab-ci.yml, Jenkinsfile, etc.)

## Environment Configuration

**Required env vars:**
- None - Application requires no environment variables

**Secrets location:**
- Not applicable - No secrets management required

## Webhooks & Callbacks

**Incoming:**
- None - No webhook endpoints

**Outgoing:**
- None - No external callbacks or webhooks initiated

## API Endpoints (Internal Only)

**Product Catalog:**
- `GET /api/products` - Returns product catalog (no external integration)

**Bundle Management:**
- `GET /api/bundles/templates` - Returns available bundle templates
- `POST /api/bundles` - Creates custom bundle
- `POST /api/bundles/clone` - Clones and modifies a template

**Discount & Notifications:**
- `POST /api/discounts/apply` - Applies discount strategy (in-memory)
- `POST /api/discounts/undo` - Reverts last discount
- `POST /api/products/{id}/subscribe` - Subscribes to price change notifications (in-memory only)

---

*Integration audit: 2026-04-07*
