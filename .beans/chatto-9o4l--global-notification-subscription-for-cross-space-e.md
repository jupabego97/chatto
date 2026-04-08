---
# chatto-9o4l
title: Push-based unread state changes
status: draft
type: feature
priority: normal
created_at: 2025-12-30T20:15:36Z
updated_at: 2026-02-08T13:33:22Z
parent: chatto-bq1a
---

Push unread state changes to clients via instance events, replacing the current pull-based model.

> **Note:** The cross-space notification subscription part of the original scope was completed in chatto-r382. The `myInstanceEvents` subscription already delivers DM, mention, and thread reply notifications. This bean now focuses on **push-based unread indicators**.

## Current State

Unread indicators work via a pull/inference model:
- Client queries `hasUnread` on rooms
- Client infers unread from incoming `MessagePostedEvent`s
- Server tracks `room_read_status` and `room_last_seq` in KV
- `markRoomAsRead()` syncs read position when entering a room

This works for single-device use but is fragile for multi-device sync (read state from one device is not pushed to others).

## Proposed Change

Add an `UnreadCountChanged` event to instance events:

```
UnreadCountChanged { spaceId, roomId, hasUnread }
```

Published when:
- A new message arrives in a room the user is a member of (but not currently viewing)
- A room is marked as read (from any device)

This enables:
- Multi-device sync (reading on phone clears badge on desktop)
- Simpler frontend logic (subscribe to events instead of inferring from messages)
- Server-authoritative unread state

## Implementation Tasks

- [ ] Add `UnreadStateChangedEvent` to proto event definitions (live-only, not persisted)
- [ ] Add GraphQL type and union membership
- [ ] Publish event when `markRoomAsRead()` is called
- [ ] Publish event when new message arrives for room members
- [ ] Frontend: subscribe and update unread indicators from events
- [ ] Frontend: remove inference-based unread detection

## Proto Impact

This would be a **live-only event** (not persisted to JetStream), so no risk to persisted proto stability.
