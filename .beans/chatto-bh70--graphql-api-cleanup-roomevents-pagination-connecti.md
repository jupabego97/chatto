---
# chatto-bh70
title: 'GraphQL API cleanup: roomEvents pagination connection type'
status: in-progress
type: task
priority: normal
created_at: 2026-03-24T15:09:43Z
updated_at: 2026-03-25T08:40:30Z
---

Wrap roomEvents return type in RoomEventsConnection with hasOlder/hasNewer metadata. Requires core changes to compute pagination bounds and frontend updates to use new fields instead of heuristics.

## Plan

- [x] Core: Add RoomEventsResult struct, update GetRoomEvents and GetRoomEventsAfter to return it
- [x] GraphQL schema: Add RoomEventsConnection type, update roomEvents return type
- [x] Codegen: Run mise codegen
- [x] GraphQL resolver: Adapt to new core return type
- [x] Frontend: Update all query definitions and consumption sites
- [x] Tests: Verify Go tests and e2e pass

## Summary of Changes

Wrapped the `roomEvents` GraphQL query in a `RoomEventsConnection` type with `hasOlder`/`hasNewer` booleans, replacing frontend count-based heuristics with server-authoritative pagination metadata. PR #585.
