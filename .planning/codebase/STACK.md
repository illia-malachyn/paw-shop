# Technology Stack

**Analysis Date:** 2026-04-07

## Languages

**Primary:**
- Go 1.23 - Backend server and core business logic

**Secondary:**
- HTML/CSS - Static frontend served from `static/` directory
- JavaScript - Client-side interactions in HTML

## Runtime

**Environment:**
- Go 1.23

**Package Manager:**
- Go Modules (go.mod)
- Lockfile: Not detected (no go.sum file committed)

## Frameworks

**Core:**
- Go standard library `net/http` - HTTP server and request handling
- No external web frameworks (e.g., no Chi, Gin, Echo)

**Testing:**
- Go standard library `testing` - Built-in testing framework
- `net/http/httptest` - HTTP test utilities

**Build/Dev:**
- No build tool configuration detected (standard `go build`/`go run`)

## Key Dependencies

**Standard Library Only:**
- `net/http` - HTTP server, handlers, and request/response management
- `encoding/json` - JSON encoding/decoding for API responses
- `strings` - String manipulation (URL parsing)
- `bytes` - Byte buffer operations
- `math` - Mathematical operations (likely for discount calculations)
- `fmt` - Formatted output and logging
- `log` - Basic logging

**No Third-Party Dependencies:**
- Project uses only Go standard library
- go.mod contains only module declaration and Go version

## Configuration

**Environment:**
- Hardcoded port: `:8080` in `cmd/server/main.go`
- No environment variable configuration detected
- Static files served from `./static/` directory

**Build:**
- Standard Go toolchain
- No Makefile, build.sh, or custom build configuration

## Platform Requirements

**Development:**
- Go 1.23+ installed
- Text editor or IDE with Go support

**Production:**
- Go 1.23+ runtime OR compiled binary
- Port 8080 must be available
- `./static/` directory must be accessible relative to binary execution

---

*Stack analysis: 2026-04-07*
