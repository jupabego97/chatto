---
# chatto-f4vk
title: Add slow consumer detection for subscription channels
status: todo
type: task
priority: low
created_at: 2026-02-16T13:52:58Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzzzzw
parent: chatto-v29q
---

## Problem

Subscription event channels (e.g., in `StreamMySpaceEvents`, `StreamMyInstanceEvents`) use buffered channels. If a client is slow to consume, the channel fills up and the sending goroutine blocks. There's no detection or graceful handling of slow consumers.

## Locations

- `cli/internal/core/core.go` — `StreamMySpaceEvents` (line ~1497)
- `cli/internal/core/core.go` — `StreamMyInstanceEvents` (line ~1898)

## Recommended Approach

1. Use a non-blocking send with `select` and `default` case
2. When a send would block, either:
   a. Drop the event and increment a counter (log warning periodically)
   b. Close the channel to disconnect the slow consumer
3. Add a metric or log for slow consumer detection

## Notes

- The current blocking behavior means one slow client can back-pressure the event processing goroutine
- Option (b) is simpler and forces the client to reconnect, which is the standard pattern for real-time subscriptions
