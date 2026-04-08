---
# chatto-cktm
title: Clamp roomEvents query limit
status: todo
type: bug
priority: normal
created_at: 2026-03-17T09:38:48Z
updated_at: 2026-03-18T05:34:10Z
order: zz
parent: chatto-pmb4
---

The limit parameter on roomEvents and roomEventsAround is accepted without upper bound. A client can pass limit: 2147483647, causing server memory exhaustion. The int32-to-uint32 conversion also means negative values become very large positive values.

Fix: clamp to max 500 in both RoomEvents and RoomEventsAround resolvers.

**Files:** cli/internal/graph/query.resolvers.go (lines 47-49, ~159)
**Severity:** Medium
