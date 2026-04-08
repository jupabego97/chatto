---
# chatto-bs8g
title: Fix goroutine leak in GetUsersBatch
status: todo
type: bug
priority: normal
created_at: 2026-02-16T13:52:13Z
updated_at: 2026-03-18T05:34:10Z
order: zzk
parent: chatto-v29q
---

## Problem

`GetUsersBatch` in `users.go:195-221` spawns goroutines for each user ID lookup but doesn't check context cancellation. If the parent context is cancelled (e.g., client disconnects), the spawned goroutines continue running until they complete their KV lookups.

## Location

- `cli/internal/core/users.go` lines 195-221

## Recommended Fix

1. Check `ctx.Err()` before spawning each goroutine
2. Inside each goroutine, select on `ctx.Done()` alongside the KV get operation
3. Consider using an `errgroup.Group` with a derived context for proper cancellation propagation

## Impact

Minor — goroutines will eventually complete since KV lookups are fast, but under load with many cancelled requests this could accumulate
