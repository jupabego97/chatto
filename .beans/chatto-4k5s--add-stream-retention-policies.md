---
# chatto-4k5s
title: Add stream retention policies
status: todo
type: task
priority: high
created_at: 2026-02-16T13:50:00Z
updated_at: 2026-03-18T05:34:10Z
order: zk
parent: chatto-v29q
---

## Problem

Space event streams (`SPACE_{spaceId}_EVENTS`) have no retention policies configured — no `MaxAge`, `MaxMsgs`, or `MaxBytes`. Streams grow unbounded, leading to ever-increasing storage costs.

## Relevant Files

- `cli/internal/core/core.go:1276-1304` — `ensureSpaceStream()` creates streams without retention config

## Approach

1. Add configurable retention to `chatto.toml` (e.g., `[nats] max_age = "90d"`)
2. Apply `MaxAge` and `Discard: DiscardOld` to `StreamConfig` in `ensureSpaceStream()`
3. Use `CreateOrUpdateStream` so existing streams get updated
4. Default to a sensible value (90 days?) but allow unlimited for self-hosters

```go
_, err := c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
    // ... existing fields ...
    MaxAge:  cfg.NATS.MaxAgeOrDefault(), // e.g., 90 * 24 * time.Hour
    Discard: jetstream.DiscardOld,
})
```

## Testing

- Unit test: verify stream config includes retention
- Integration test: verify old messages are purged after MaxAge
