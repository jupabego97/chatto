---
# chatto-312s
title: Increase ephemeral consumer InactiveThreshold from 10s to 30s
status: todo
type: task
priority: normal
created_at: 2026-02-16T13:52:32Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzzk
parent: chatto-v29q
---

## Problem

Multiple ephemeral consumers in `rooms.go` use `InactiveThreshold: 10 * time.Second`. This is quite aggressive — if the consumer falls behind or the client has a brief network hiccup, the consumer gets garbage collected and the query/subscription fails.

## Locations

- `cli/internal/core/rooms.go` lines 1851, 1927, 1948, 1980, 1999, 2193

## Recommended Fix

1. Define a constant `defaultEphemeralInactiveThreshold = 30 * time.Second`
2. Replace all hardcoded `10 * time.Second` values with this constant
3. Consider making this configurable via `chatto.toml` for production tuning

## Notes

- 30s is the NATS default for ephemeral consumers and is well-tested in production environments
- The 10s value provides no meaningful benefit over 30s since these consumers are explicitly deleted when done
