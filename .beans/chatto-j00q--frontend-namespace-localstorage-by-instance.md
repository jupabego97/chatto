---
# chatto-j00q
title: 'Frontend: Namespace localStorage by instance'
status: draft
type: task
priority: normal
created_at: 2026-03-03T11:46:54Z
updated_at: 2026-03-03T11:49:30Z
parent: chatto-wadw
blocked_by:
    - chatto-1uxi
---

Scope all localStorage keys by instance ID so that data from different instances doesn't collide.

## Context

Currently all localStorage keys use a flat `chatto:` prefix:
- `chatto:lastRooms` — last visited room per space
- `chatto:lastSpace` — last visited space
- `chatto:threadPaneWidth` — thread pane width
- `chatto:preferences` — notification sound preference
- `chatto:recentReactions` — recently used emoji reactions
- `chatto:pinnedRoomIds` — pinned room IDs

If a user connects to two instances, space IDs and room IDs could collide.

## Requirements

- [ ] Change key format from `chatto:{key}` to `chatto:{instanceId}:{key}` for instance-specific data
- [ ] Keep truly global preferences (like `threadPaneWidth`, UI preferences) un-namespaced
- [ ] Migrate existing data: treat existing `chatto:` keys as belonging to the "home" instance
- [ ] Create a helper function `instanceKey(instanceId, key)` used by all storage utilities
- [ ] Update all files that read/write localStorage
- [ ] Write tests for the migration and key construction

## Which keys are instance-scoped vs global

**Instance-scoped** (different per instance):
- `chatto:lastRooms` → `chatto:{instanceId}:lastRooms`
- `chatto:lastSpace` → `chatto:{instanceId}:lastSpace`
- `chatto:pinnedRoomIds` → `chatto:{instanceId}:pinnedRoomIds`
- `chatto:recentReactions` → `chatto:{instanceId}:recentReactions` (different emoji sets per instance possible)
- `space:{spaceId}:collapsed-sections` → `chatto:{instanceId}:space:{spaceId}:collapsed-sections`

**Global** (same regardless of instance):
- `chatto:threadPaneWidth` — UI layout preference
- `chatto:preferences` — notification sound, etc.

## Key files
- `frontend/src/lib/storage/lastRoom.ts`
- `frontend/src/lib/storage/threadPaneWidth.ts`
- `frontend/src/lib/state/userPreferences.svelte.ts`
- `frontend/src/lib/state/recentReactions.svelte.ts`
- `frontend/src/lib/RoomList.svelte` (pinned rooms)

## Blocked by
- Instance registry (need instance IDs)
