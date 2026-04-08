---
# chatto-mvja
title: Use FlushTimeout instead of Flush for NATS publishes
status: todo
type: bug
priority: normal
created_at: 2026-02-16T13:50:27Z
updated_at: 2026-03-18T05:34:10Z
order: zzz
parent: chatto-v29q
---

## Problem

At `core.go:1082,1107,1127`, `nc.Flush()` is called after publishing events. `Flush()` blocks indefinitely until the server acknowledges — on a network partition this hangs forever.

## Relevant Files

- `cli/internal/core/core.go:1082` — after `publishSpaceEvent`
- `cli/internal/core/core.go:1107` — after `publishSpaceLiveEvent`
- `cli/internal/core/core.go:1127` — after `publishInstanceEvent`

## Fix

Replace `c.nc.Flush()` with `c.nc.FlushTimeout(5 * time.Second)` at all three locations. Handle the timeout error appropriately (log warning, don't fail the operation).

## Testing

Existing tests should continue passing. No new tests needed unless we want to simulate network partition.
