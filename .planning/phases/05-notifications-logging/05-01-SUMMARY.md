---
phase: 05-notifications-logging
plan: 01
subsystem: notify
tags: [facade, pattern, notifications, go]
dependency_graph:
  requires: []
  provides: [internal/notify]
  affects: []
tech_stack:
  added: []
  patterns: [Facade]
key_files:
  created:
    - internal/notify/facade.go
    - internal/notify/facade_test.go
  modified: []
decisions:
  - Used io.Writer for ConsoleNotifier and FileNotifier to allow testing without real I/O
  - errors.Join aggregates partial failures so no channel error is silently dropped
metrics:
  duration_minutes: 15
  completed: "2026-04-07T22:30:47Z"
  tasks_completed: 2
  files_changed: 2
---

# Phase 05 Plan 01: Facade Pattern — NotificationFacade Summary

**One-liner:** NotificationFacade hiding ConsoleNotifier + FileNotifier complexity behind NotifyUser and NotifyOrderStatusChanged using stdlib errors.Join for error aggregation.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Implement Facade pattern | eceff7a | internal/notify/facade.go |
| 2 | Unit tests for Facade pattern | eceff7a | internal/notify/facade_test.go |

## What Was Built

- `Notifier` interface with `Notify(userID, message string) error`
- `ConsoleNotifier` struct — writes `[CONSOLE] User {id}: {msg}\n` to an `io.Writer`
- `FileNotifier` struct — writes `[FILE] User {id}: {msg}\n` to an `io.Writer`
- `NotificationFacade` struct — holds `[]Notifier`, iterates all notifiers on each call
- `NewNotificationFacade(console, file io.Writer)` constructor
- `NotifyUser(userID, message string) error` — sends through all notifiers, aggregates errors
- `NotifyOrderStatusChanged(orderID, status string) error` — formats and delegates to NotifyUser

## Decisions Made

- **io.Writer for notifiers:** Both ConsoleNotifier and FileNotifier accept `io.Writer` rather than using `os.Stdout`/real files directly, enabling clean unit testing with `bytes.Buffer`.
- **errors.Join for aggregation:** Errors from individual notifiers are collected and joined so partial failures are visible to callers (Pitfall 10 compliance — errors not swallowed).
- **Slice of Notifier interface:** Facade holds `[]Notifier` enabling easy extension with additional channels without changing facade methods.

## Deviations from Plan

None — plan executed exactly as written.

## Test Coverage

- `TestNotifyUser_SendsToBothChannels` — verifies both console and file buffers receive `[CONSOLE]`/`[FILE]` prefixed output with userID and message
- `TestNotifyOrderStatusChanged_FormatsMessage` — verifies orderID and status appear in output
- `TestFacade_ErrorPropagation` — uses a `failingWriter` that returns error on Write, verifies facade returns non-nil error

All 3 tests pass. Full `go test ./...` passes with no regressions.

## Self-Check: PASSED

- [x] `internal/notify/facade.go` exists
- [x] `internal/notify/facade_test.go` exists
- [x] Commit eceff7a exists and contains both files
- [x] `go test ./internal/notify/ -v` all pass
- [x] `go vet ./internal/notify/` no issues
