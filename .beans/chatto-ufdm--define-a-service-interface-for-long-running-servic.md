---
# chatto-ufdm
title: Define a Service interface for long-running services
status: draft
type: task
created_at: 2026-01-04T12:01:56Z
updated_at: 2026-01-04T12:01:56Z
---

Establish a common interface for services that:
- Can be started and managed by errgroup
- Might in the future live in external processes (e.g., KMS server)

## Current State

Two different patterns exist:
- HTTP server: `Run(ctx) error` - blocks until ctx cancelled
- NATS: `Start(ctx, errgroup)` - returns immediately, registers shutdown goroutine

## Proposed Interface

```go
// Service represents a long-running service that can be started and stopped.
// Run blocks until ctx is cancelled or an error occurs.
type Service interface {
    Run(ctx context.Context) error
}
```

## Considerations

- NATS is special: other services need `*server.Server` for in-process connections
- Options for handling NATS dependency:
  1. Return server from a separate method after construction
  2. Pass server via config/DI after NATS is ready
  3. Keep NATS as a special case (must be ready before others start)

## Context

This came up during refactoring of `start.go` to use errgroup for service coordination.