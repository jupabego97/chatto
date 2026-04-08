---
# chatto-012a
title: 'Frontend: Multi-instance sidebar'
status: in-progress
type: feature
priority: normal
created_at: 2026-03-03T11:47:36Z
updated_at: 2026-03-18T05:31:56Z
order: zw
parent: chatto-wadw
blocked_by:
    - chatto-46kr
---

Redesign the SpaceList sidebar to show instance groups with their space icons. This is the primary user-facing change — the "unified inbox" sidebar.

## Context

The current SpaceList (`frontend/src/lib/SpaceList.svelte`) is a single column of:
- DM icon
- Space icons (from one instance)
- Create Space / Browse / Admin buttons
- User avatar

For multi-instance, this becomes a grouped list where each instance shows its own spaces.

## Requirements

- [ ] Redesign SpaceList to show instance groups
- [ ] Each instance group shows: instance icon/avatar, then its space icons underneath
- [ ] Instance icons are clickable (expanding/collapsing the instance's spaces, or selecting it)
- [ ] Unread indicators on instance icons (aggregate of all spaces' unreads)
- [ ] Unread indicators on individual space icons (as today)
- [ ] Connection status indicator per instance (connected/disconnected/connecting)
- [ ] "Add Instance" button at the bottom of the instance list
- [ ] DM icon per instance (if the instance supports DMs)
- [ ] Space ordering within each instance (as configured per instance)
- [ ] Instance ordering (drag to reorder? or chronological by add date?)
- [ ] Active space highlight works across instances
- [ ] Clicking a space from a different instance switches the active instance
- [ ] Write e2e tests for multi-instance sidebar interactions

## Visual Design

```
┌──────────┐
│ [Inst 1] │  ← Instance icon (with unread dot if any space has unreads)
│  [DM]    │  ← DM icon for this instance
│  [Sp A]  │  ← Space icons for this instance
│  [Sp B]  │
│          │
│ [Inst 2] │  ← Second instance
│  [DM]    │
│  [Sp C]  │
│          │
│ [+]      │  ← Add Instance
│          │
│ [⚙]     │  ← Global settings
│ [👤]    │  ← User avatar (or remove if per-instance?)
└──────────┘
```

## Implementation Notes

### Data flow
The sidebar needs to:
1. Read the instance registry for the list of instances
2. For each instance, read its space memberships (from the instance's GraphQL client)
3. For each instance, read its unread counts (from the instance's unread store)
4. Aggregate into a grouped list

### SpaceList refactor
The current `SpaceList.svelte` fetches `SpaceListInit` query on mount. For multi-instance, each instance needs its own query. Consider:
- A `InstanceSpaceList` component that renders one instance's spaces
- The parent `SpaceList` iterates over instances and renders `InstanceSpaceList` for each

### Active space tracking
Currently `activeSpaceId` comes from the URL params. For multi-instance, we also need `activeInstanceId`. Clicking a space sets both.

### Key files
- `frontend/src/lib/SpaceList.svelte` — major refactor
- `frontend/src/lib/SpaceIcon.svelte` — may need instance awareness
- `frontend/src/routes/chat/+layout.svelte` — passes `activeSpaceId`, needs `activeInstanceId`

### Blocked by
- Instance registry (need instance list)
- GraphQL client factory (need per-instance clients to fetch space lists)
- Instance-scoped state management (need per-instance unread stores)

### Does NOT include
- "Add Instance" flow UI (separate bean)
- Route restructuring (separate bean)
