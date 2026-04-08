---
# chatto-ow6h
title: 'Frontend: Route restructuring for multi-instance'
status: draft
type: feature
priority: normal
created_at: 2026-03-03T11:48:19Z
updated_at: 2026-03-03T11:49:30Z
parent: chatto-wadw
blocked_by:
    - chatto-1uxi
    - chatto-301x
    - chatto-46kr
---

Add an instance segment to the URL hierarchy so that routes are scoped per instance. Enables deep linking to a specific instance's space/room.

## Context

Current routes:
```
/chat/[spaceId]/[roomId]
/chat/dm/[conversationId]
/chat/admin/...
```

These assume a single instance. For multi-instance, we need to know which instance a space belongs to.

## Requirements

- [ ] Add instance segment to all chat routes: `/i/[instanceId]/chat/[spaceId]/[roomId]`
- [ ] `instanceId` is the local slug from the instance registry (e.g., "home", "work")
- [ ] Update all `goto()` and `resolve()` calls to include the instance segment
- [ ] Update all `$page.params` destructuring to include `instanceId`
- [ ] The root `/chat` route redirects to the last active instance
- [ ] Deep links to `/i/work/chat/abc123/def456` activate the correct instance and navigate to the space/room
- [ ] Handle unknown instance IDs gracefully (show "instance not found" or prompt to add)
- [ ] Backward compatibility: `/chat/[spaceId]` without instance segment works for the home instance (redirect to `/i/home/chat/[spaceId]`)
- [ ] Update all route guards (auth checks, membership checks) to use the correct instance's client
- [ ] Write e2e tests for cross-instance navigation

## Route Structure

```
/                                          — Landing/login
/i/[instanceId]/chat                       — Instance chat view (redirect to last space)
/i/[instanceId]/chat/spaces                — Browse spaces on this instance
/i/[instanceId]/chat/[spaceId]             — Space view
/i/[instanceId]/chat/[spaceId]/[roomId]    — Room view
/i/[instanceId]/chat/[spaceId]/[roomId]/[threadId]
/i/[instanceId]/chat/[spaceId]/admin/...   — Space admin
/i/[instanceId]/chat/dm/...                — DMs on this instance
/i/[instanceId]/chat/admin/...             — Instance admin
/instances/add                             — Add instance flow
/settings/...                              — Global client settings
```

## Implementation Notes

### SvelteKit route groups
Use SvelteKit route groups to nest under `/i/[instanceId]/`:
```
src/routes/
  i/
    [instanceId]/
      chat/
        +layout.svelte    — sets active instance, provides instance client via context
        [spaceId]/
          +layout.svelte  — validates space membership on this instance
          [roomId]/
            ...
```

### Layout responsibilities
The `[instanceId]/+layout.svelte` layout:
1. Reads `instanceId` from params
2. Looks up the instance in the registry
3. Sets it as the active instance (updates active client, state)
4. Provides the instance's GraphQL client via Svelte context
5. Renders children

### Migration strategy
This is a big routing change. Consider:
1. Create the new route structure alongside the old one
2. Add redirects from old routes to new ones (`/chat/abc → /i/home/chat/abc`)
3. Remove old routes once everything works

### Key files
- All files under `frontend/src/routes/chat/` — move to `frontend/src/routes/i/[instanceId]/chat/`
- All navigation helpers, `goto()` calls, `resolve()` calls throughout the frontend
- `frontend/src/lib/SpaceList.svelte` — links
- `frontend/src/lib/storage/lastRoom.ts` — needs instance awareness

### Blocked by
- Instance registry (need instance IDs for routing)
- GraphQL client factory (need per-instance clients)
- Instance-scoped state management (layout sets active instance state)

### Risk
This is one of the highest-effort beans because it touches nearly every file that does navigation. Consider doing it in stages:
1. First, add the `/i/[instanceId]/` wrapper without changing existing routes
2. Then migrate routes one section at a time (chat, admin, settings, dm)
