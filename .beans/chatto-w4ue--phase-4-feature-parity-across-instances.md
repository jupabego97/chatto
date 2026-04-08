---
# chatto-w4ue
title: 'Phase 4: Feature parity across instances'
status: todo
type: task
priority: normal
created_at: 2026-03-12T16:36:52Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzzw
parent: chatto-e88o
blocked_by:
    - chatto-t4nr
---

## Goal

Make features that currently only work on the home instance also work on remote instances.

## Already Multi-Instance (no changes needed)

- GraphQL client manager (home=cookies, remote=bearer tokens — correct)
- Event bus architecture (home via context, remote via manager)
- Last room/space storage (namespaced by instance ID)
- Instance permissions/roles
- Session channel (global logout is correct)
- Auth/session handling (home redirects to login, remote clears — correct)
- User settings (global is fine for now)

## Needs Work

### DMs on remote instances
- `startDMWith()` hardcoded to `homeClient`
- `DMConversationList` queries home only
- DM routes only exist under home instance (`/chat/-/dm/*`)
- Requires: pass instanceId through DM functions, update routing
- **Blocker:** Backend must support DM spaces per instance

### Push notifications per instance
- `subscribe()`/`unsubscribe()` hardcoded to `homeClient`
- Service worker tied to home origin
- Requires: accept instanceId param, route to correct client
- **Decision:** One subscription per instance, or home-only push?

### Notification sync & badge aggregation
- `NotificationSync.svelte` only listens to home instance events
- PWA badge reflects home notifications only
- Requires: aggregate across instances or track per active instance
- **Decision:** Global badge (sum all instances) or per-active-instance?

### Presence tracking on active instance
- `initPresenceTracking()` called only with `homeClient`
- User presence only reported to home instance
- Requires: call for each active instance, not just home
- **Easiest fix** — just pass the active instance's client

### AppHeader (notification bell, MOTD)
- Renders above `[instanceId]` route tree, hardcoded to home
- Bell shows home notifications only; MOTD shows home MOTD
- Requires: move inside route tree OR derive from active instance
- **Decision:** Should header reflect active instance or always home?

## Suggested Order

1. Presence tracking (small, self-contained)
2. AppHeader context-awareness (moderate, visible impact)
3. Notification aggregation (moderate, important UX)
4. DMs on remote instances (large, needs backend work)
5. Push notifications per instance (large, needs service worker changes)
