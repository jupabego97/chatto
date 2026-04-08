<script lang="ts">
  import { createSpaceEventBus, startSpaceSubscription } from '$lib/spaceEventBus.svelte';
  import { usePresenceChange, useReconnectCallback } from '$lib/hooks';
  import { useConnection } from '$lib/state/instance/connection.svelte';
  import { getPresenceCache } from '$lib/state/presenceCache.svelte';
  import type { Snippet } from 'svelte';

  let { spaceId, children }: { spaceId: string; children: Snippet } = $props();

  // Create event bus context synchronously
  const spaceEventBus = createSpaceEventBus();

  // Capture presence cache during init (context must be read synchronously)
  const presenceCache = getPresenceCache();

  const connection = useConnection();

  // Start space event subscription (messages, room events, reactions, presence).
  // Explicitly track reconnectCount so the subscription restarts after WebSocket
  // reconnections — don't rely solely on graphql-ws to re-subscribe, which can
  // silently fail if the subscription was in an intermediate state during the drop.
  $effect(() => {
    const conn = connection();
    void conn.reconnectCount;
    return startSpaceSubscription(spaceEventBus, conn.client, spaceId);
  });

  // Clear presence cache after WebSocket reconnection
  useReconnectCallback(() => {
    console.log('WebSocket reconnected, clearing presence cache');
    presenceCache.clear();
  });

  // Populate global presence cache from space events so that any UserAvatar
  // (including newly-mounted ones like popovers) sees the latest presence.
  usePresenceChange((userId, status) => {
    presenceCache.update(userId, status);
  });
</script>

<div data-testid="space-subscription-active" class="hidden"></div>
{@render children()}
