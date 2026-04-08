---
# chatto-3ewy
title: Simplify RUNTIME bucket by deriving state from JetStream
status: draft
type: feature
priority: low
created_at: 2026-01-18T16:01:58Z
updated_at: 2026-01-18T16:01:58Z
---

Investigate removing derivable data from `SPACE_{id}_RUNTIME` now that event IDs in subjects enable O(1) JetStream lookups via `GetLastMsgForSubject`.

## Context

PR #118 added event IDs to message subjects, enabling O(1) lookups. PR #119 extended this to threading. The `SPACE_{id}_RUNTIME` bucket was created before these optimizations existed, storing precomputed values that may now be derivable directly from JetStream.

## Current RUNTIME Contents

The bucket stores four types of data:

| Key Pattern | Purpose | Data | Derivable? |
|-------------|---------|------|------------|
| `room_first_seq.{roomId}` | Pagination boundary detection | uint64 (string) | ✅ Yes |
| `room_last_seq.{roomId}` | Unread detection (room side) | uint64 (binary) | ✅ Yes |
| `room_last_msg_at.{roomId}` | Sort rooms by activity | int64 nanos (binary) | ✅ Yes |
| `room_read_status.{userId}.{roomId}` | User's read cursor | uint64 (binary) | ❌ No |

### Detailed Analysis

**1. `room_first_seq.{roomId}`** - First event sequence in room
- Written once per room (on first message)
- Used to detect "reached beginning" during pagination
- **Can derive from**: Stream's first message for subject `space.{s}.room.{r}.>`
- **Current code**: `setRoomFirstSequence()`, `getRoomFirstSequence()` in `rooms.go`

**2. `room_last_seq.{roomId}`** - Last event sequence in room
- Written on every message post
- Core of unread detection: compare against user's read cursor
- **Can derive from**: `GetLastMsgForSubject("space.{s}.room.{r}.>")` returns sequence
- **Current code**: `setRoomLastSequence()`, `GetRoomLastSequence()` in `rooms.go`
- **Hot path**: Called on every `HasUnread()` check

**3. `room_last_msg_at.{roomId}`** - Timestamp of last message
- Written on every message post
- Used to sort DM conversations and room lists by recent activity
- **Can derive from**: Last message's `CreatedAt` field (from JetStream metadata or event payload)
- **Current code**: `setRoomLastMessageAt()`, `GetRoomLastMessageAt()` in `rooms.go`

**4. `room_read_status.{userId}.{roomId}`** - User's read position
- Written when user marks room as read
- User-initiated state, not event-derived
- **Cannot be derived** - this is user preference, not audit data
- **Must remain in KV**

## Trade-offs

### Benefits of Removing Derivable Keys

1. **Single source of truth**: Stream is authoritative, no dual-write synchronization
2. **Smaller KV footprint**: Fewer keys per room
3. **Simpler writes**: `PostMessage` only writes to stream + BODIES, not RUNTIME
4. **No stale data risk**: Can't have desync between stream and cached values

### Costs of Deriving from Stream

1. **Latency**: `GetLastMsgForSubject` may be slower than KV `Get()` (needs benchmarking)
2. **Parsing overhead**: May need to unpack message to get timestamp (though JetStream metadata includes time)
3. **API differences**: JetStream returns message metadata differently than KV
4. **Unread detection hot path**: `HasUnread()` is called frequently; must stay fast

### Hybrid Approach Consideration

Could keep `room_last_seq` in KV for hot-path unread detection but remove the others. Or use in-memory caching with stream as source of truth.

## Benchmarking Required

Before implementing, benchmark these operations:

```go
// Current approach (KV)
kv.Get("room_last_seq.{roomId}")  // ~50µs typical

// Proposed approach (stream)
stream.GetLastMsgForSubject("space.{s}.room.{r}.>")  // needs measurement
```

The Direct Get API (`$JS.API.DIRECT.GET`) should be fast, but we need to verify it's acceptable for the `HasUnread()` hot path.

## Tasks

### Phase 1: Benchmark
- [ ] Add benchmark tests comparing KV get vs GetLastMsgForSubject latency
- [ ] Measure with varying room sizes (10, 100, 1000, 10000 messages)
- [ ] Test under load (concurrent requests)

### Phase 2: Implement (if benchmarks acceptable)
- [ ] Create `getRoomLastSequenceFromStream()` helper
- [ ] Create `getRoomFirstSequenceFromStream()` helper
- [ ] Create `getRoomLastMessageAtFromStream()` helper
- [ ] Add feature flag or config to switch between KV and stream-derived
- [ ] Update `PostMessage` to skip RUNTIME writes for derivable fields
- [ ] Update `HasUnread()` to use stream-derived sequence

### Phase 3: Cleanup
- [ ] Remove `setRoomFirstSequence()`, `setRoomLastSequence()`, `setRoomLastMessageAt()`
- [ ] Simplify RUNTIME bucket description in ARCHITECTURE.md
- [ ] Update tests

### Phase 4: Documentation
- [ ] Update ARCHITECTURE.md KV bucket documentation
- [ ] Document the decision and trade-offs

## Decision Points

1. **All or nothing?** Remove all three derivable keys, or keep `room_last_seq` for hot path?
2. **Caching layer?** Add in-memory cache for frequently accessed sequences?
3. **Migration?** Existing RUNTIME data can be ignored (will be re-derived), or cleaned up?

## Files to Modify

- `cli/internal/core/rooms.go` - Primary changes (getters/setters)
- `cli/internal/core/core.go` - Bucket configuration (if simplifying)
- `ARCHITECTURE.md` - Documentation updates

## Notes

- The `room_read_status` key pattern MUST remain - it's user state, not derivable
- Consider renaming bucket to `SPACE_{id}_USER_STATE` if only read status remains
- Breaking change is acceptable (early stage)
