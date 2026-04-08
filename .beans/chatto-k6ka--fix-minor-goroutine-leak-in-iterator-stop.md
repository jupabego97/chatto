---
# chatto-k6ka
title: Fix minor goroutine leak in iterator stop
status: todo
type: bug
priority: low
created_at: 2026-02-16T13:52:32Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzzzzk
parent: chatto-v29q
---

## Problem

In `core.go:1678-1685`, a goroutine is spawned to call `iter.Stop()` but it has no timeout or context check. If `iter.Stop()` blocks (e.g., due to a stuck NATS connection), this goroutine will leak.

## Location

- `cli/internal/core/core.go` lines 1678-1685

## Recommended Fix

Wrap the `iter.Stop()` call in a context with timeout:

```go
go func() {
    stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    select {
    case <-stopCtx.Done():
        // iter.Stop() is stuck, log and abandon
        logger.Warn("iterator stop timed out")
    default:
        iter.Stop()
    }
}()
```

## Impact

Very minor — `iter.Stop()` rarely blocks, but this is good defensive programming
