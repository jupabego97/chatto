---
# chatto-gvp2
title: Extract room data into shared space-scoped store
status: draft
type: feature
created_at: 2026-01-21T17:20:11Z
updated_at: 2026-01-21T17:20:11Z
---

## Motivation

Currently room data is managed independently by multiple components:

1. **RoomList.svelte** - queries `me.roomMemberships(spaceId)` for sidebar
2. **RoomDirectory.svelte** - queries `space.rooms` + `me.roomMemberships` for browse/discovery
3. **DMConversationList.svelte** - queries DM space rooms separately
4. **SpaceList.svelte** - queries DM unread status

**Problems:**
- Duplicate queries when multiple components need room data
- Full `loadRooms()` reloads on join/leave events instead of targeted updates
- `RoomDeletedEvent` and `RoomUpdatedEvent` are defined but never handled
- Race condition handling exists in RoomList but not RoomDirectory
- No shared source of truth for room state

## Proposed Solution

Create `spaceRooms.svelte.ts` - a shared context store scoped to a single space, similar to the `roomMembers.svelte.ts` pattern.

### Store Design

```typescript
type SpaceRoom = {
  id: string;
  name: string;
  description?: string | null;
  hasUnread: boolean;
  hasMention: boolean;
  isMember: boolean;  // Whether current user is a member
};

type SpaceRoomsState = {
  rooms: SpaceRoom[];
  loading: boolean;
};

// Context API
createSpaceRooms(): SpaceRoomsStore
getSpaceRoomsState(): SpaceRoomsState
getJoinedRooms(): SpaceRoom[]      // Filter: isMember === true
getAllRooms(): SpaceRoom[]          // All rooms (for browse)
getRoomById(id: string): SpaceRoom | undefined
```

### Event Handling

The store should handle these events with targeted updates:

| Event | Update |
|-------|--------|
| `UserJoinedRoomEvent` | Add room or set `isMember = true` |
| `UserLeftRoomEvent` | Set `isMember = false` (don't remove - still browseable) |
| `MessagePostedEvent` | Set `hasUnread = true` for room |
| `MentionNotificationEvent` | Set `hasMention = true` for room |
| `RoomCreatedEvent` | Add new room to list |
| `RoomDeletedEvent` | Remove room from list |
| `RoomUpdatedEvent` | Update room name/description |

### Scope Boundary

- **In scope**: Regular spaces with rooms
- **Out of scope**: DM space (has different data shape with participants, keep separate)

## Implementation Plan

### Phase 1: Create the store

- [ ] Create `frontend/src/lib/state/spaceRooms.svelte.ts`
- [ ] Define `SpaceRoom` type with all fields needed by consumers
- [ ] Implement `createSpaceRooms()` with context setup
- [ ] Implement getters: `getSpaceRoomsState()`, `getJoinedRooms()`, `getAllRooms()`
- [ ] Add update methods: `setRooms()`, `addRoom()`, `removeRoom()`, `updateRoom()`, `markUnread()`, `markMention()`

### Phase 2: Wire up in space layout

- [ ] Call `createSpaceRooms()` in `[spaceId]/+layout.svelte`
- [ ] Create a unified query that fetches rooms with all needed fields
- [ ] Subscribe to space events and update store with targeted changes
- [ ] Handle `RoomCreatedEvent`, `RoomDeletedEvent`, `RoomUpdatedEvent`

### Phase 3: Migrate RoomList

- [ ] Remove local room loading from `RoomList.svelte`
- [ ] Use `getJoinedRooms()` from store
- [ ] Remove local event handlers (store handles them)
- [ ] Keep optimistic unread/mention state merge logic in store

### Phase 4: Migrate RoomDirectory

- [ ] Remove local room loading from `RoomDirectory.svelte`
- [ ] Use `getAllRooms()` from store
- [ ] Filter by `!isMember` for joinable rooms display
- [ ] Remove local event handlers

### Phase 5: Update instance event handling

- [ ] Move `MentionNotificationEvent` handling to store (currently in +layout.svelte)
- [ ] Ensure store marks `hasMention` correctly

### Phase 6: Clean up and test

- [ ] Remove unused queries after migration
- [ ] Run `mise codegen-frontend`
- [ ] Run `mise test-frontend` - all tests should pass
- [ ] Manual testing: join/leave rooms, create rooms, delete rooms, unread indicators

## Files Affected

**New:**
- `frontend/src/lib/state/spaceRooms.svelte.ts`

**Modified:**
- `frontend/src/routes/chat/[spaceId]/+layout.svelte` - create store, wire events
- `frontend/src/lib/RoomList.svelte` - consume store instead of local state
- `frontend/src/lib/RoomDirectory.svelte` - consume store instead of local state
- `frontend/src/lib/gql/` - generated types will change

**Not modified:**
- `frontend/src/lib/dm/DMConversationList.svelte` - DM space stays separate

## Considerations

### Different data needs
RoomList needs `hasUnread`/`hasMention`, RoomDirectory needs `description`. Store should include superset of all fields.

### Permission boundary
`canBrowseRooms` permission gates visibility - this is handled at layout level and should continue to work.

### Race conditions
Store should implement optimistic state merging (preserve `hasUnread`/`hasMention` if either server or local state says true).

### Query efficiency
Single query on mount that fetches all room data vs. multiple smaller queries. Trade-off: slightly more data upfront, but simpler and more consistent.

## Testing

- E2E: Room creation appears in list without refresh
- E2E: Joining room moves it to "joined" section
- E2E: Leaving room removes from sidebar (if not in room browse view)
- E2E: Unread indicators update correctly
- E2E: Mention indicators update correctly
- E2E: Room rename updates in all views