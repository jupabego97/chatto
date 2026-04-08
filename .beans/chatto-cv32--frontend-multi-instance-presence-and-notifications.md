---
# chatto-cv32
title: 'Frontend: Multi-instance presence and notifications'
status: draft
type: feature
priority: normal
created_at: 2026-03-03T11:48:57Z
updated_at: 2026-03-03T11:49:30Z
parent: chatto-wadw
blocked_by:
    - chatto-301x
    - chatto-46kr
    - chatto-lk24
---

Handle presence reporting and notification aggregation across multiple connected instances.

## Context

### Presence
Currently a module-level singleton (`presenceTracking.ts`) reports ONLINE/AWAY status to one backend via `updateMyPresence` mutation. For multi-instance, presence needs to be reported to each connected instance.

### Notifications
- In-app notifications come via the instance event bus and are stored in `notificationStore` (singleton)
- Push notifications are registered via a single Web Push subscription to one backend's VAPID key
- The app badge shows a single unread count

## Requirements

### Presence
- [ ] Report presence to all connected instances (not just the active one)
- [ ] Idle detection and visibility tracking remain global (shared across all instances)
- [ ] When user goes AWAY, all instances are notified
- [ ] When an instance disconnects, its presence reporting stops gracefully
- [ ] Write tests for multi-instance presence

### Notifications
- [ ] Aggregate notification counts across all instances for the app badge
- [ ] In-app notification list shows notifications from all instances, grouped by instance
- [ ] Each notification indicates which instance it came from
- [ ] Clicking a notification navigates to the correct instance/space/room
- [ ] Push notifications: register with each instance's VAPID key separately (each instance gets its own push subscription)
- [ ] Service worker handles push from multiple instances (use the `data.instanceUrl` or similar to route clicks)
- [ ] Write tests for notification aggregation

## Implementation Notes

### Presence tracking refactor
```typescript
// Instead of one global instance, track presence per instance
class MultiInstancePresence {
  #instances = new Map<string, { client: GraphQLClient; status: PresenceStatus }>();
  
  // Called by the global idle/visibility detector
  updateAll(status: PresenceStatus) {
    for (const [id, inst] of this.#instances) {
      inst.client.mutation(UpdateMyPresence, { status });
    }
  }
}
```

### Push notification registration
Each Chatto instance has its own VAPID key pair. The service worker needs to register a separate push subscription for each instance. This may require:
- Multiple `pushManager.subscribe()` calls with different `applicationServerKey` values
- Storing a mapping of VAPID key → instance URL in the service worker's state
- OR: accepting that push only works for one instance (the home instance) initially

**Recommendation**: Start with push notifications only for the home instance. Multi-instance push is complex and can be deferred.

### Notification aggregation
```typescript
class AggregatedNotifications {
  // Notifications from all instances, sorted by time
  all = $derived(this.#instances.flatMap(i => i.notifications).sort(byTime));
  unreadCount = $derived(this.all.filter(n => !n.read).length);
}
```

### Key files
- `frontend/src/lib/presenceTracking.ts` — refactor for multi-instance
- `frontend/src/lib/state/notifications.svelte.ts` — aggregate across instances
- `frontend/src/lib/notifications/pushNotifications.ts` — per-instance registration
- `frontend/src/service-worker.ts` — multi-instance push handling
- `frontend/src/lib/notifications/appBadge.ts` — aggregate counts

### Blocked by
- GraphQL client factory (need per-instance mutation clients)
- Instance-scoped state management (notifications per instance)
- Multi-instance event buses (notifications arrive via event bus)
