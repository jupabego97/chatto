---
# chatto-lk24
title: 'Frontend: Multi-instance event buses'
status: in-progress
type: feature
priority: normal
created_at: 2026-03-03T11:48:36Z
updated_at: 2026-03-18T05:31:56Z
order: zs
parent: chatto-wadw
blocked_by:
    - chatto-301x
    - chatto-46kr
---

Create per-instance event bus subscriptions so that each connected instance receives real-time events (messages, presence, notifications, unreads).

## Context

Currently there are two event bus systems:
1. **Instance Event Bus** (`instanceEventBus.svelte.ts`) — module-level singleton, subscribes to `myInstanceEvents` for space CRUD, profile changes, notification sync, unread markers
2. **Space Event Bus** (`spaceEventBus.svelte.ts`) — context-based, one per active space, subscribes to `mySpaceEvents(spaceId)` for messages, presence, typing

Both use the singleton GraphQL client. For multi-instance, each instance needs its own event subscriptions.

## Requirements

- [ ] Create an instance event bus per connected instance (for unread counts, notifications)
- [ ] Space event bus already context-scoped — make it use the active instance's client
- [ ] For non-active instances, maintain lightweight subscriptions (instance events only, not space events)
- [ ] When switching active instance, tear down old space event bus and create new one
- [ ] Unread count updates from all instances flow to the sidebar
- [ ] Notification events from all instances are aggregated
- [ ] Connection lifecycle: start subscriptions when instance is added, stop when removed
- [ ] Handle reconnection per-instance (one instance going offline shouldn't affect others)
- [ ] Write tests for event routing

## Architecture

```
Instance 1 (active):
  ├─ Instance Event Bus (full: unreads, notifications, space CRUD)
  ├─ Space Event Bus for Space A (full: messages, presence, typing)
  └─ Space Event Bus for Space B (if visible — e.g., in sidebar unread tracking)

Instance 2 (inactive):
  └─ Instance Event Bus (lightweight: unreads, notifications only)
      No space event buses (saves WebSocket bandwidth)
```

## Implementation Notes

### Event bus per instance
Each `InstanceState` (from the state management bean) owns its own instance event bus:
```typescript
class InstanceState {
  eventBus: InstanceEventBus;  // subscribes to myInstanceEvents via this instance's client
  // ...
}
```

### Space event bus
The space event bus is already context-scoped (`createSpaceEventBus()` in the space layout). It just needs to use the active instance's GraphQL client instead of the global singleton.

### Sidebar unread aggregation
The sidebar needs unread counts from ALL instances. Each instance's event bus dispatches unread updates to a global aggregator:
```typescript
class GlobalUnreadAggregator {
  // Combines unread counts from all instance event buses
  totalUnread = $derived(/* sum across instances */);
  getInstanceUnread(id: string): number { ... }
}
```

### Key files
- `frontend/src/lib/instanceEventBus.svelte.ts` — refactor from singleton to per-instance
- `frontend/src/lib/spaceEventBus.svelte.ts` — accept client as parameter
- `frontend/src/lib/state/roomUnread.svelte.ts` — per-instance unread tracking

### Blocked by
- GraphQL client factory (need per-instance WebSocket connections)
- Instance-scoped state management (event buses live in InstanceState)

### Does NOT include
- Multi-instance presence tracking (separate bean)
- Multi-instance push notifications (separate bean)
