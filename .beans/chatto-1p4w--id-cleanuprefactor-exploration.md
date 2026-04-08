---
# chatto-1p4w
title: ID cleanup/refactor exploration
status: draft
type: task
priority: normal
tags:
    - backend
    - nats
    - exploration
created_at: 2025-12-07T20:54:47Z
updated_at: 2025-12-14T21:37:48Z
parent: chatto-12xq
---

Optimize internal ID representation by using uint64 subject sequence IDs internally while keeping NanoIDs for external APIs.

## Motivation

- **Wire size**: uint64 varints (1-10 bytes, typically 1-3) vs 14-byte NanoID strings (~16 bytes with length prefix)
- **Performance**: Integer comparisons are single CPU instructions vs string comparisons
- **Consistency**: Messages already use uint64 sequence IDs; this unifies the approach

## Architecture Overview

### ID Types

| Entity  | External ID (GraphQL) | Internal ID (Proto) |
|---------|----------------------|---------------------|
| User    | NanoID (14 chars)    | uint64 subject seq  |
| Space   | NanoID (14 chars)    | uint64 subject seq  |
| Room    | NanoID (14 chars)    | uint64 subject seq  |
| Message | uint64 sequence      | uint64 sequence     |

### Subject Sequence IDs

NATS JetStream provides subject-scoped sequence IDs. When publishing to a subject, NATS returns a monotonically increasing sequence number scoped to that specific subject.

**Strategy**: Dedicated subjects for entity registration, where the subject sequence becomes the permanent internal ID.

- `instance.users.register` → subject sequence = user internal ID
- `instance.spaces.create` → subject sequence = space internal ID  
- `space.{spaceInternalId}.rooms.create` → subject sequence = room internal ID (space-scoped)

### ID Mapping Storage

**Instance-level KV bucket: `ID_MAPS`**

```
user.ext.{nanoId}     → uint64 (internal ID)
user.int.{uint64}     → nanoId (external ID)
space.ext.{nanoId}    → uint64 (internal ID)
space.int.{uint64}    → nanoId (external ID)
```

**Space-level KV bucket: `SPACE_{spaceInternalId}_ID_MAPS`**

```
room.ext.{nanoId}     → uint64 (internal ID)
room.int.{uint64}     → nanoId (external ID)
```

### Caching Strategy

ID mappings are immutable once created. Implement in-memory caches:
- Bounded LRU cache for external→internal lookups (hot path on every request)
- Bounded LRU cache for internal→external lookups (hot path on every response)
- Populate on first access, never invalidate (mappings never change)
- Consider populating on startup for small deployments

---

## Implementation Plan

### Phase 1: ID Mapping Infrastructure

Create the foundation for ID translation without changing existing behavior.

- [ ] Create `ID_MAPS` KV bucket in instance setup
- [ ] Create `SPACE_{id}_ID_MAPS` KV bucket in space setup
- [ ] Implement `IDMapper` service with methods:
  - `RegisterUser(nanoId) -> uint64` - publishes to registration subject, stores mapping
  - `RegisterSpace(nanoId) -> uint64`
  - `RegisterRoom(spaceInternalId, nanoId) -> uint64`
  - `UserExternalToInternal(nanoId) -> uint64`
  - `UserInternalToExternal(uint64) -> nanoId`
  - (same pattern for spaces/rooms)
- [ ] Implement LRU caching layer for ID lookups
- [ ] Add metrics for cache hit/miss rates
- [ ] Write tests for IDMapper service

### Phase 2: Proto Schema Changes

Update protobuf definitions to use uint64 for internal IDs.

- [ ] Create new proto field versions with uint64 types:
  ```protobuf
  message Space {
    uint64 id = 10;           // NEW: internal ID
    string external_id = 11;  // NEW: NanoID for API
    // Deprecate: string id = 1;
    string name = 2;
    string description = 3;
  }
  ```
- [ ] Update all event protos to use uint64 for entity references:
  ```protobuf
  message RoomCreatedEvent {
    uint64 space_id = 10;     // NEW
    uint64 room_id = 11;      // NEW
    // Keep string fields temporarily for migration
    string name = 3;
    string description = 4;
  }
  ```
- [ ] Update `Event` wrapper if needed
- [ ] Run `mise codegen` and fix compilation errors

### Phase 3: Core Layer - Entity Creation

Update entity creation to use the new ID system.

- [ ] Update `CreateUser`:
  1. Generate NanoID
  2. Call `IDMapper.RegisterUser(nanoId)` → get uint64
  3. Store user with uint64 as primary ID, NanoID in `external_id` field
  4. Update KV key pattern: `user.{uint64}` instead of `user.{nanoId}`
  5. Update email index to point to uint64
- [ ] Update `CreateSpace` (same pattern)
- [ ] Update `CreateRoom` (same pattern, space-scoped)
- [ ] Update all KV key patterns in ARCHITECTURE.md

### Phase 4: Core Layer - Entity Lookups & References

Update all code that references entities by ID.

- [ ] Update `GetUser`, `GetSpace`, `GetRoom` to accept uint64
- [ ] Update membership operations to use uint64 IDs
- [ ] Update message operations (space_id, room_id references)
- [ ] Update reaction operations
- [ ] Update presence operations
- [ ] Update role/permission operations
- [ ] Audit all KV key patterns and update

### Phase 5: GraphQL Translation Layer

Add ID translation at the API boundary.

- [ ] Create `ExternalID` and `InternalID` types for type safety
- [ ] Update input types to accept string IDs (NanoIDs)
- [ ] Add translation in resolver input handling:
  ```go
  func (r *mutationResolver) CreateRoom(ctx context.Context, spaceID string, ...) {
    internalSpaceID, err := r.idMapper.SpaceExternalToInternal(spaceID)
    // ... use internalSpaceID internally
  }
  ```
- [ ] Add translation in resolver output:
  ```go
  func (r *spaceResolver) ID(ctx context.Context, obj *pb.Space) (string, error) {
    return r.idMapper.SpaceInternalToExternal(obj.Id)
  }
  ```
- [ ] Update all resolvers that handle entity IDs
- [ ] Update GraphQL schema if needed (IDs remain `ID!` type)

### Phase 6: Data Migration

Migrate existing data to the new ID scheme.

- [ ] Write migration script that:
  1. Reads all existing users, spaces, rooms
  2. Generates uint64 IDs (starting from 1, incrementing)
  3. Creates ID mappings in new KV buckets
  4. Rewrites entity data with new ID fields
  5. Rewrites all events with uint64 IDs (or marks as legacy)
  6. Updates all KV keys to new patterns
- [ ] Test migration on copy of production data
- [ ] Plan rollback strategy

### Phase 7: Cleanup

Remove deprecated code paths.

- [ ] Remove deprecated string ID fields from protos
- [ ] Remove any dual-ID handling code
- [ ] Update documentation
- [ ] Performance benchmark: compare before/after

---

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Migration data loss | Test on copy first, implement rollback |
| Cache memory usage | Bounded LRU caches with configurable size |
| ID mapping KV failure | Mappings are immutable; can rebuild from entity data |
| Breaking API changes | External API unchanged (still uses NanoIDs) |

## Open Questions

1. **Room ID scope**: Should room uint64 IDs be globally unique or space-scoped? Space-scoped is more efficient but requires (spaceId, roomId) tuples everywhere.

2. **Message body IDs**: Currently use NanoID (`message_body_id`). Should these also become uint64? They're only used for KV key lookups.

3. **Existing data**: Full migration vs dual-read (check both old and new key patterns)?

## References

- NATS JetStream subject sequences: messages have both stream sequence and subject sequence
- Current proto definitions: `proto/chatto/v1/`
- Current architecture: `ARCHITECTURE.md`