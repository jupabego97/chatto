---
# chatto-nzv1
title: Unified instance store
status: draft
type: task
created_at: 2026-03-08T16:48:12Z
updated_at: 2026-03-08T16:48:12Z
---

Consolidate the three separate per-instance stores (instanceRegistry, graphqlClientManager, instanceEventBusManager) into a single unified InstanceStore that exposes a `ConnectedInstance` object per instance with direct references to its client, event bus, etc.

Currently every multi-instance view has to import all three managers and correlate them by instance ID. A unified store would let views simply iterate `instanceStore.instances` and access `inst.client` directly.

Proposed shape:
```ts
interface ConnectedInstance {
  id: string;
  url: string;
  name: string;
  isHome: boolean;
  client: Client;
  eventBus: EventBus;
}
```

Wait until we have more multi-instance views to confirm the right shape before implementing.
