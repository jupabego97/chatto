---
# chatto-bq1a
title: Improve read/unread markers system
status: draft
type: epic
priority: normal
created_at: 2026-01-18T23:18:53Z
updated_at: 2026-02-08T13:32:12Z
---

Improve the read/unread marker system to be more robust, support threads, and support multi-device sync.

> **Note:** The notification system (bell icon, notification center, DM/mention/reply notifications, real-time delivery) was completed in chatto-r382. This epic now focuses solely on **read/unread markers**.

## Current State

Room-level unread tracking works via KV entries:
- `room_read_status.{userId}.{roomId}` - User's last read sequence position
- `room_last_seq.{roomId}` - Room's latest message sequence
- `HasUnread()` compares these: `room_last_seq > user_last_read_seq`

The frontend calls `markRoomAsRead()` when entering a room, and subscribes to events for real-time unread indicators.

## Remaining Problems

### 1. No thread-level unread tracking
Current system only tracks room-level read position. No way to:
- Show unread badge on thread indicators
- Track "followed threads" vs all threads
- Know if specific thread has new replies

### 2. Pull-based vs push-based unread state
Current model requires clients to query/infer unread state. A push-based model (chatto-9o4l) where the server publishes `UnreadCountChanged` events would be more reliable for multi-device sync.

### 3. Reactions and unread state
- Reactions in DMs don't trigger unread indicator (chatto-s1du)
- Design decision needed: should reactions trigger unread?

## Proposed Phases

### Phase 1: Push-based unread updates (chatto-9o4l)
- Server pushes `UnreadCountChanged { spaceId, roomId, count }` via instance events
- Client subscribes and updates local unread state
- Enables reliable multi-device sync

### Phase 2: Thread-level unread tracking (chatto-q48e)
- Add `thread_read_status.{userId}.{rootId}` KV entries
- Track which threads user has "followed" (auto-follow on reply, or explicit)
- Show unread badge on thread indicators

## Design Decisions Needed

- [ ] Should reactions trigger unread? (DMs only? All rooms? Configurable?)
- [ ] Should unread count be exact or boolean?
- [ ] Should "followed threads" be implicit (replied to) or explicit (subscribe button)?

## Related Beans
- chatto-s1du: DM reactions don't trigger unread indicator
- chatto-q48e: Thread unread tracking
- chatto-9o4l: Push-based unread state changes
- chatto-3ewy: Simplify RUNTIME bucket
- chatto-7af2: Thread Following

## Child Features
- chatto-q48e: Thread unread tracking
- chatto-9o4l: Push-based unread state changes
- chatto-2a9q: Notification Preferences (user settings)
- chatto-v9rm: Room Muting (suppress notifications per-room)
- chatto-7af2: Thread Following (follow/unfollow threads)
