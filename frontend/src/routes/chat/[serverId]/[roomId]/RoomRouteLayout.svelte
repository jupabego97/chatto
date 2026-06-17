<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { getActiveServer } from '$lib/state/activeServer.svelte';
  import { serverRegistry } from '$lib/state/server/registry.svelte';
  import type { ResolvedPathname } from '$app/types';
  import {
    roomMessagePathForSegment,
    roomPathForSegment,
    roomThreadPathForSegment,
    type RoomRouteKind
  } from '$lib/roomUrls';
  import { clearLastRoom } from '$lib/storage/lastRoom';
  import Room from './Room.svelte';

  let { data, children, routeKind }: {
    data: { roomId?: string };
    children?: import('svelte').Snippet;
    routeKind: 'legacy-id' | 'name';
  } = $props();

  let roomSegment = $derived(data.roomId);

  const activeServerId = $derived(getActiveServer());
  const stores = $derived(serverRegistry.getStore(activeServerId));

  // Wait for the active server's merged rooms store (channels + DMs) to
  // settle before letting children mount. Without this, a freshly-loaded
  // room page can fire queries against the URL roomId before the store has
  // decided whether the room exists, briefly showing the not-found redirect.
  const roomsStore = $derived(stores.rooms);
  const ready = $derived(!roomsStore.isInitialLoading);

  let threadId = $derived(page.params.threadId);
  let messageId = $derived(page.params.messageId);
  let resolvedRoomId = $state<string | null>(null);
  let resolveRequest = 0;

  const isMessageLinkMode = $derived(/\/m\/[^/]+$/.test(page.url.pathname));

  function canonicalRoomPath(
    canonicalSegment: string,
    canonicalRouteKind: RoomRouteKind
  ): ResolvedPathname {
    let path: ResolvedPathname;
    if (isMessageLinkMode && messageId) {
      path = roomMessagePathForSegment(
        page.params.serverId!,
        canonicalSegment,
        messageId,
        canonicalRouteKind
      );
    } else if (threadId) {
      path = roomThreadPathForSegment(
        page.params.serverId!,
        canonicalSegment,
        threadId,
        canonicalRouteKind
      );
    } else {
      path = roomPathForSegment(page.params.serverId!, canonicalSegment, canonicalRouteKind);
    }
    return `${path}${page.url.search}${page.url.hash}` as ResolvedPathname;
  }

  $effect(() => {
    if (!ready || !roomSegment) {
      resolvedRoomId = null;
      return;
    }

    const request = ++resolveRequest;
    stores.roomRoutes.resolve(roomSegment, routeKind).then((resolved) => {
      if (request !== resolveRequest) return;
      resolvedRoomId = resolved?.roomId ?? null;
      if (!resolved) {
        clearLastRoom(activeServerId);
        goto(resolve('/chat/[serverId]', { serverId: page.params.serverId! }), { replaceState: true });
        return;
      }
      if (resolved.canonicalSegment === roomSegment && resolved.canonicalRouteKind === routeKind) return;

      goto(canonicalRoomPath(resolved.canonicalSegment, resolved.canonicalRouteKind), { replaceState: true });
    });
  });
</script>

{#if ready && resolvedRoomId}
  {#if isMessageLinkMode}
    <!-- Message link resolver: renders +page.svelte which fetches + redirects -->
    {@render children?.()}
  {:else}
    <!--
      Room is rendered in the layout so it stays mounted when navigating
      between room and thread URLs. This prevents unnecessary reloads.
    -->
    {#key activeServerId}
      <Room roomId={resolvedRoomId} {threadId} />
    {/key}
  {/if}
{/if}
